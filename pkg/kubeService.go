package kubernetescontroller

//annotations aliyun/clb_id=xxx aliyun/clb_port=xxx
import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var CacheUid types.UID

func kubeConfig() (config *rest.Config) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join("kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}
	return config
}

func Service() {
	config := kubeConfig()
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error()) // 适当处理错误
	}
	watchService(clientset)
}

func GetNodes(clientset *kubernetes.Clientset) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	NodesList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return []string{}, fmt.Errorf("error listing nodes: %s", err)
	}
	nodes := []string{}
	for _, v := range NodesList.Items {
		for _, address := range v.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				nodes = append(nodes, address.Address)
				Logger("info", fmt.Sprintf("NodeExternalIP %s,%s", address.Address, address.Type))
			} else {
				Logger("info", fmt.Sprintf("NodeExternalIP %s,%s", address.Address, address.Type))
				nodes = append(nodes, address.Address)
			}
		}
	}
	return nodes, nil
}

// checkStringInSlice 检查一个字符串是否在切片中
func checkStringInSlice(Name string, Services []corev1.Service) bool {
	for _, v := range Services {
		if v.Name == Name {
			return true
		}
	}
	return false
}

func watchService(clientset *kubernetes.Clientset) {
	//start int
	isInitialEvent := true
	services, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	//end
	watch, err := clientset.CoreV1().Services("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		Logger("error", fmt.Sprintf("Error watching pods: %s", err))
		return
	}
	defer watch.Stop()
	for event := range watch.ResultChan() {
		service, ok := event.Object.(*corev1.Service)
		if !ok {
			Logger("error", fmt.Sprintf("Watch Unexpected type %s", event.Type))
			continue
		}
		code := checkStringInSlice(service.Name, services.Items)
		// start first skip
		if isInitialEvent && code {
			continue
		} else {
			isInitialEvent = false
		}
		if service.Spec.Type != corev1.ServiceTypeNodePort {
			Logger("info", fmt.Sprintf("service type is not NodePort,%s,%s,%s", service.Spec.Type, service.Name, event.Type))
			continue
		}
		clb_id, exists := service.Annotations["aliyun/clb_id"]
		if !exists {
			Logger("error", fmt.Sprintf("service clb_id is not exists %s,%s", service.Name, event.Type))
			ServiceUid := service.UID
			err = DeleteService(clientset, service.Namespace, service.Name, ServiceUid)
			if err != nil {
				panic(err)
			}
			continue
		}
		vpc_id, exists := service.Annotations["aliyun/vpc_id"]
		if !exists {
			Logger("error", fmt.Sprintf("service vpc_id is not exists %s,%s", service.Name, event.Type))
			ServiceUid := service.UID
			err = DeleteService(clientset, service.Namespace, service.Name, ServiceUid)
			if err != nil {
				panic(err)
			}
			continue
		}
		regionId, exists := service.Annotations["aliyun/regionid"]
		if !exists {
			Logger("error", fmt.Sprintf("service regionId is not exists %s,%s", service.Name, event.Type))
			ServiceUid := service.UID
			err = DeleteService(clientset, service.Namespace, service.Name, ServiceUid)
			if err != nil {
				panic(err)
			}
			continue
		}
		CLBPort, exists := service.Annotations["aliyun/clb_port"]
		if !exists {
			Logger("error", fmt.Sprintf("service clb_port is not exists %s,%s", service.Name, event.Type))
			ServiceUid := service.UID
			_err := DeleteService(clientset, service.Namespace, service.Name, ServiceUid) //rollback Service
			if _err != nil {
				panic(_err)
			}
			continue
		}
		IntClbPort, err := strconv.Atoi(CLBPort)
		if err != nil {
			Logger("error", fmt.Sprintf("service clb_port to int failed %s,%s,%s", err, service.Name, CLBPort))
			ServiceUid := service.UID
			_err := DeleteService(clientset, service.Namespace, service.Name, ServiceUid)
			if _err != nil {
				panic(err)
			}
			continue
		}
		clb_port := int32(IntClbPort)
		switch event.Type {
		case "ADDED":
			ServiceUid := service.UID                                  //get service uid
			nodeString, exists := service.Annotations["aliyun/ecs_ip"] //get ecs ip if not exists get nodes
			// nodes := []string{}
			var nodes []string
			if !exists {
				nodes, err = GetNodes(clientset)
				if err != nil {
					_err := DeleteService(clientset, service.Namespace, service.Name, ServiceUid) //rollback Service
					if _err != nil {
						panic(err)
					}
				}
			} else {
				nodes = strings.Split(nodeString, ",")
			}

			Logger("info", fmt.Sprintf("Added %s %s %s %s %d %s %d", regionId, service.Namespace, service.Name, clb_id, clb_port, vpc_id, service.Spec.Ports[0].NodePort))
			AliyunSLB(regionId, clb_id, clb_port, vpc_id, service.Namespace, service.Name, service.Spec.Ports[0].NodePort, string(service.Spec.Ports[0].Protocol), nodes, "Create", clientset, ServiceUid)
		case "DELETED":
			ServiceUid := service.UID //get service uid
			if CacheUid == ServiceUid {
				continue
			}
			Logger("info", fmt.Sprintf("Deleted %s %s %s %s %d %s %d", regionId, service.Namespace, service.Name, clb_id, clb_port, vpc_id, service.Spec.Ports[0].NodePort))
			AliyunSLB(regionId, clb_id, clb_port, vpc_id, service.Namespace, service.Name, service.Spec.Ports[0].NodePort, string(service.Spec.Ports[0].Protocol), []string{}, "Delete", clientset, ServiceUid)
		case "MODIFIED":
			Logger("info", fmt.Sprintf("MODIFIED %s %s %s %s %d %s %d", regionId, service.Namespace, service.Name, clb_id, clb_port, vpc_id, service.Spec.Ports[0].NodePort))
			continue
		case "BOOKMARK":
			Logger("info", fmt.Sprintf("BOOKMARK %s %s %s %s %d %s %d", regionId, service.Namespace, service.Name, clb_id, clb_port, vpc_id, service.Spec.Ports[0].NodePort))
			continue
		}
	}
}

func DeleteService(clientset *kubernetes.Clientset, namespace string, ServiceName string, ServiceUid types.UID) error {
	if CacheUid == ServiceUid {
		CacheUid = "" //set null
		return nil
	}
	err := clientset.CoreV1().Services(namespace).Delete(context.TODO(), ServiceName, metav1.DeleteOptions{})
	CacheUid = ServiceUid
	return err
}

func AliyunSLB(regionId string, clb_id string, clb_port int32, vpc_id string, namespace string, ServiceName string, BackendPort int32, Protocol string, ServerIps []string, Method string, clientset *kubernetes.Clientset, ServiceUid types.UID) {
	aliyunclb := Aliyunclb{}
	aliyunclb.regionId = &regionId
	aliyunclb.loadBalancerId = &clb_id
	aliyunclb.ListenerPort = &clb_port
	aliyunclb.Protocol = &Protocol
	aliyunclb.VPCId = &vpc_id
	VServerGroupName := fmt.Sprintf("vServerGroup-%s-%s-%d-%d", namespace, ServiceName, BackendPort, clb_port)
	aliyunclb.VServerGroupName = &VServerGroupName
	Description := fmt.Sprintf("CLB-%s-%s-%d-%d", namespace, ServiceName, BackendPort, clb_port)
	aliyunclb.Description = &Description
	VServerGroupBackendSpecs, err := AliyunEcs(ServerIps, 100, BackendPort, regionId, "ecs", vpc_id)
	if err != nil {
		panic(err)
	}
	aliyunclb.VServerGroupBackendSpec = VServerGroupBackendSpecs
	aliyunclb._main()
	if Method == "Create" {
		err = aliyunclb.CreateSLBListener()
		if err != nil {
			errorHandler(err)
			err = DeleteService(clientset, namespace, ServiceName, ServiceUid) //rollback
			panic(err)
		}

	} else if Method == "Delete" {
		err = aliyunclb.DeleteSLBListener()
		if err != nil {
			errorHandler(err)
		}
	} else {
		return
	}

}
