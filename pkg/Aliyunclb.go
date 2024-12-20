package kubernetescontroller

import (
	"encoding/json"
	"fmt"

	env "github.com/alibabacloud-go/darabonba-env/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	openapiv2 "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	slb "github.com/alibabacloud-go/slb-20140515/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyunclb struct {
	regionId                *string
	loadBalancerId          *string
	ListenerPort            *int32
	VServerGroupName        *string
	vServerGroupId          *string
	Protocol                *string
	Status                  *string
	Description             *string
	client                  *slb.Client
	VPCId                   *string
	VServerGroupBackendSpec []VServerGroupBackendSpec
}

type VServerGroupBackendSpec struct {
	Weight      int32
	ServerId    string
	Port        int32
	Type        string
	ServerIp    string
	Description string
}

func (r *Aliyunclb) Initialization() (_result *slb.Client) {
	config := &openapi.Config{}
	// 您的AccessKey ID
	config.AccessKeyId = env.GetEnv(tea.String("ACCESS_KEY_ID"))
	// 您的AccessKey Secret
	config.AccessKeySecret = env.GetEnv(tea.String("ACCESS_KEY_SECRET"))
	// 您的可用区ID
	config.RegionId = r.regionId
	endpoint := fmt.Sprintf("slb.%s.aliyuncs.com", *r.regionId)
	config.Endpoint = &endpoint
	// SLB 创建比较慢，所以超时设置为 10 秒
	config.ReadTimeout = tea.Int(10000)
	_result = &slb.Client{}
	_result, _err := slb.NewClient(config)
	if _err != nil {
		panic(_err)
	}
	return _result
}
func (r *Aliyunclb) _main() {
	// 区域ID
	// 初始化请求参数配置client。
	client := r.Initialization()
	r.client = client
}

// Protocol string, Port int32, regionId string, loadBalancerId string
func (r *Aliyunclb) QuerySLBListenerx() (err error) {
	client := r.client
	Protocol := *r.Protocol
	if Protocol == "TCP" {
		describeLoadBalancerTCPListenerAttributeRequest := &slb.DescribeLoadBalancerTCPListenerAttributeRequest{}
		describeLoadBalancerTCPListenerAttributeRequest.RegionId = r.regionId
		describeLoadBalancerTCPListenerAttributeRequest.LoadBalancerId = r.loadBalancerId
		describeLoadBalancerTCPListenerAttributeRequest.ListenerPort = r.ListenerPort
		describeLoadBalancerTCPListenerAttributeResponse, _err := client.DescribeLoadBalancerTCPListenerAttribute(describeLoadBalancerTCPListenerAttributeRequest)
		if _err != nil {
			errorHandler(_err)
			Logger("error", fmt.Sprintf("error QuerySLBListenerx %s %s", err, _err))
			return _err
		}
		loadBalancerTCPListenerAttribute := describeLoadBalancerTCPListenerAttributeResponse.Body
		Logger("info", fmt.Sprintf("querySLBListener TCP %v %v", loadBalancerTCPListenerAttribute.Status, loadBalancerTCPListenerAttribute.VServerGroupId))
		r.Status = loadBalancerTCPListenerAttribute.Status
		r.vServerGroupId = loadBalancerTCPListenerAttribute.VServerGroupId
		return nil
	} else if Protocol == "UDP" {
		describeLoadBalancerUDPListenerAttributeRequest := &slb.DescribeLoadBalancerUDPListenerAttributeRequest{}
		describeLoadBalancerUDPListenerAttributeRequest.RegionId = r.regionId
		describeLoadBalancerUDPListenerAttributeRequest.LoadBalancerId = r.loadBalancerId
		describeLoadBalancerUDPListenerAttributeRequest.ListenerPort = r.ListenerPort
		describeLoadBalancerUDPListenerAttributeResponse, _err := client.DescribeLoadBalancerUDPListenerAttribute(describeLoadBalancerUDPListenerAttributeRequest)
		if _err != nil {
			return _err
		}
		loadBalancerUDPListenerAttribute := describeLoadBalancerUDPListenerAttributeResponse.Body
		Logger("info", fmt.Sprintf("querySLBListener UDP %v %v", loadBalancerUDPListenerAttribute.Status, loadBalancerUDPListenerAttribute.VServerGroupId))
		r.Status = loadBalancerUDPListenerAttribute.Status
		r.vServerGroupId = loadBalancerUDPListenerAttribute.VServerGroupId
		return nil
	}
	return fmt.Errorf("protocol not match %s", Protocol)

}

func (r *Aliyunclb) QuerySLBListener(deleteQuery bool) error {
	client := r.client
	var loadBalancerIds []*string = []*string{}
	// describeLoadBalancerListenersRequest := &slb20140515.DescribeLoadBalancerListenersRequest{}
	describeLoadBalancerListenersRequest := &slb.DescribeLoadBalancerListenersRequest{}
	describeLoadBalancerListenersRequest.ListenerProtocol = r.Protocol
	describeLoadBalancerListenersRequest.LoadBalancerId = append(loadBalancerIds, r.loadBalancerId)
	describeLoadBalancerListenersRequest.RegionId = r.regionId
	// describeLoadBalancerListenersRequest.ListenerPort = r.ListenerPort
	DescribeLoadBalancerListeners, err := client.DescribeLoadBalancerListeners(describeLoadBalancerListenersRequest)
	if err != nil {
		errorHandler(err)
		return fmt.Errorf("QuerySLBListener failed %s", err)
	}
	for _, listener := range DescribeLoadBalancerListeners.Body.Listeners {
		if *listener.ListenerPort == *r.ListenerPort {
			if deleteQuery {
				// if delete server group
				r.vServerGroupId = listener.VServerGroupId
				return nil
			} else {
				return fmt.Errorf("QuerySLBListener is exist Port Staus %v,Port %v", *listener.Status, *listener.ListenerPort)
			}

		}
	}
	return nil

}
func UpdateSLBListener() {
}

// Protocol string, Port int32, regionId string, loadBalancerId string
func (r *Aliyunclb) CreateSLBListener() (err error) {
	client := r.client
	Protocol := *r.Protocol
	err = r.CreateSLBVServerGroup()
	if err != nil {
		return fmt.Errorf("createSLBVServerGroup failed %s", err)
	}
	if Protocol == "TCP" {
		createLoadBalancerTCPListenerRequest := &slb.CreateLoadBalancerTCPListenerRequest{}
		createLoadBalancerTCPListenerRequest.LoadBalancerId = r.loadBalancerId
		createLoadBalancerTCPListenerRequest.ListenerPort = r.ListenerPort
		createLoadBalancerTCPListenerRequest.Bandwidth = tea.Int32(100)
		createLoadBalancerTCPListenerRequest.RegionId = r.regionId
		createLoadBalancerTCPListenerRequest.HealthCheckType = tea.String("tcp")
		createLoadBalancerTCPListenerRequest.Description = r.Description
		createLoadBalancerTCPListenerRequest.HealthCheckInterval = tea.Int32(10)
		createLoadBalancerTCPListenerRequest.VServerGroupId = r.vServerGroupId
		createLoadBalancerTCPListenerResponse, _err := client.CreateLoadBalancerTCPListener(createLoadBalancerTCPListenerRequest)
		if _err != nil {
			return _err
		}
		Logger("info", fmt.Sprintf("CreateLoadBalancerTCPListener success %s,%d", *createLoadBalancerTCPListenerResponse.Body.RequestId, *r.ListenerPort))
		err = r.StartSLBListener()
		if err != nil {
			return fmt.Errorf("startSLBListener failed %s,%d", err, &r.ListenerPort)
		}
		return nil
	} else if Protocol == "UDP" {
		createLoadBalancerUDPListenerRequest := &slb.CreateLoadBalancerUDPListenerRequest{}
		createLoadBalancerUDPListenerRequest.LoadBalancerId = r.loadBalancerId
		createLoadBalancerUDPListenerRequest.ListenerPort = r.ListenerPort
		createLoadBalancerUDPListenerRequest.Bandwidth = tea.Int32(100)
		createLoadBalancerUDPListenerRequest.RegionId = r.regionId
		createLoadBalancerUDPListenerRequest.HealthCheckConnectPort = r.ListenerPort
		createLoadBalancerUDPListenerRequest.Description = r.Description
		createLoadBalancerUDPListenerRequest.HealthCheckInterval = tea.Int32(10)
		createLoadBalancerUDPListenerRequest.VServerGroupId = r.vServerGroupId
		createLoadBalancerUDPListenerResponse, _err := client.CreateLoadBalancerUDPListener(createLoadBalancerUDPListenerRequest)
		if _err != nil {
			return _err
		}
		Logger("info", fmt.Sprintf("CreateLoadBalancerUDPListener success %s,%d", *createLoadBalancerUDPListenerResponse.Body.RequestId, *r.ListenerPort))
		err = r.StartSLBListener()
		if err != nil {
			return fmt.Errorf("startSLBListener failed %s,%d", err, &r.ListenerPort)
		}
		return nil
	}
	return fmt.Errorf("protocol not match %s", Protocol)
}

// port int32, regionId string, Protocol string, loadBalancerId string
func (r *Aliyunclb) StartSLBListener() error {
	client := r.client
	StartLoadBalancerListenerRequest := &slb.StartLoadBalancerListenerRequest{}
	StartLoadBalancerListenerRequest.ListenerPort = r.ListenerPort
	StartLoadBalancerListenerRequest.RegionId = r.regionId
	StartLoadBalancerListenerRequest.ListenerProtocol = r.Protocol
	StartLoadBalancerListenerRequest.LoadBalancerId = r.loadBalancerId
	StartLoadBalancerListenerResponse, _err := client.StartLoadBalancerListener(StartLoadBalancerListenerRequest)
	if _err != nil {
		return _err
	}
	Logger("info", fmt.Sprintf("startLoadBalancerListener success %s,%d", *StartLoadBalancerListenerResponse.Body.RequestId, *r.ListenerPort))
	return nil

}

// port int32, regionId string, Protocol string, loadBalancerId string
func (r *Aliyunclb) StopSLBListener() error {
	client := r.client
	StopLoadBalancerListenerRequest := &slb.StopLoadBalancerListenerRequest{}
	StopLoadBalancerListenerRequest.ListenerPort = r.ListenerPort
	StopLoadBalancerListenerRequest.RegionId = r.regionId
	StopLoadBalancerListenerRequest.ListenerProtocol = r.Protocol
	StopLoadBalancerListenerRequest.LoadBalancerId = r.loadBalancerId
	StopLoadBalancerListenerResponse, _err := client.StopLoadBalancerListener(StopLoadBalancerListenerRequest)
	if _err != nil {
		return _err
	}
	Logger("info", fmt.Sprintf("stopLoadBalancerListener success %s,%d", *StopLoadBalancerListenerResponse.Body.RequestId, *r.ListenerPort))
	return nil
}

// regionId string, Port int32, loadBalancerId string, Protocol string
func (r *Aliyunclb) DeleteSLBListener() (err error) {
	client := r.client
	//查询SLBListener
	err = r.QuerySLBListener(true)
	if err != nil {
		return fmt.Errorf("QuerySLBListener failed %v", err)
	}
	err = r.StopSLBListener()
	if err != nil {
		return fmt.Errorf("disableSLBListener failed %s", err)
	}
	DeleteLoadBalancerListenerRequest := &slb.DeleteLoadBalancerListenerRequest{}
	DeleteLoadBalancerListenerRequest.ListenerPort = r.ListenerPort
	DeleteLoadBalancerListenerRequest.ListenerProtocol = r.Protocol
	DeleteLoadBalancerListenerRequest.RegionId = r.regionId
	DeleteLoadBalancerListenerRequest.LoadBalancerId = r.loadBalancerId
	DeleteLoadBalancerListenerResponse, _err := client.DeleteLoadBalancerListener(DeleteLoadBalancerListenerRequest)
	if _err != nil {
		return _err
	}
	Logger("info", fmt.Sprintf("deleteLoadBalancerListener success %s,%d", *DeleteLoadBalancerListenerResponse.Body.RequestId, *r.ListenerPort))

	_err = r.DeleteSLBVServerGroup()
	if _err != nil {
		return _err
	}
	return nil
}

// regionId string, loadBalancerId string, VServerGroupName string
func (r *Aliyunclb) CreateSLBVServerGroup() (err error) {
	client := r.client

	err = r.QuerySLBListener(false)
	if err != nil {
		return fmt.Errorf("QuerySLBListener failed %s", err)
	}
	//add vServerGroup
	createVServerGroupRequest := &slb.CreateVServerGroupRequest{}
	createVServerGroupRequest.LoadBalancerId = r.loadBalancerId
	createVServerGroupRequest.RegionId = r.regionId
	createVServerGroupRequest.VServerGroupName = r.VServerGroupName
	BackendServersJson, err := json.Marshal(r.VServerGroupBackendSpec)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %s", err.Error())
	}
	BackendServers := string(BackendServersJson)
	createVServerGroupRequest.BackendServers = &BackendServers
	// createVServerGroupRequest.BackendServers = tea.String("[{ \"ServerId\": \"eni-xxxxxxxxx\", \"Weight\": \"100\", \"Type\": \"eni\", \"ServerIp\": \"192.168.**.**\", \"Port\":\"80\",\"Description\":\"test-112\" }]")
	createVServerGroupResponse, _err := client.CreateVServerGroup(createVServerGroupRequest)
	if _err != nil {
		return _err
	}
	r.vServerGroupId = createVServerGroupResponse.Body.VServerGroupId
	Logger("info", fmt.Sprintf("createVServerGroup success %s,%s", createVServerGroupResponse.Body, BackendServers))
	return nil
}

// vServerGroupId string, regionId string
func (r *Aliyunclb) DeleteSLBVServerGroup() (err error) {
	client := r.client
	//删除vServerGroup
	deleteVServerGroupRequest := &slb.DeleteVServerGroupRequest{}
	deleteVServerGroupRequest.VServerGroupId = r.vServerGroupId
	deleteVServerGroupRequest.RegionId = r.regionId
	DeleteVServerGroupResponse, _err := client.DeleteVServerGroup(deleteVServerGroupRequest)
	if _err != nil {
		return _err
	}
	Logger("info", fmt.Sprintf("deleteVServerGroup success %s,%s", DeleteVServerGroupResponse.Body, *r.vServerGroupId))
	return nil
}

func InitializationEcs(regionId string) (_result *ecs20140526.Client) {
	config := &openapiv2.Config{}
	// 您的AccessKey ID
	config.AccessKeyId = env.GetEnv(tea.String("ACCESS_KEY_ID"))
	// 您的AccessKey Secret
	config.AccessKeySecret = env.GetEnv(tea.String("ACCESS_KEY_SECRET"))
	// 您的可用区ID
	config.RegionId = &regionId
	endpoint := fmt.Sprintf("ecs.%s.aliyuncs.com", regionId)
	config.Endpoint = &endpoint
	// SLB 创建比较慢，所以超时设置为 10 秒
	config.ReadTimeout = tea.Int(10000)
	_result = &ecs20140526.Client{}
	_result, _err := ecs20140526.NewClient(config)
	if _err != nil {
		panic(_err)
	}
	return _result
}
func AliyunEcs(ServerIps []string, Weight int32, backendPort int32, regionId string, Type string, VPCId string) (VServerGroupBackendSpecs []VServerGroupBackendSpec, err error) {
	client := InitializationEcs(regionId)
	// ServerIpsArray := make([]*string, len(ServerIps))
	var ServerIpsArray []*string = []*string{}
	for _, ip := range ServerIps {
		ServerIpsArray = append(ServerIpsArray, &ip)
	}
	DescribeNetworkInterfacesRequest := &ecs20140526.DescribeNetworkInterfacesRequest{}
	DescribeNetworkInterfacesRequest.RegionId = &regionId
	DescribeNetworkInterfacesRequest.PrivateIpAddress = ServerIpsArray
	DescribeNetworkInterfacesRequest.VpcId = &VPCId
	// DescribeNetworkInterfaces, _err := client.DescribeNetworkInterfaces(DescribeNetworkInterfacesRequest)
	runtime := &util.RuntimeOptions{}
	DescribeNetworkInterfaces, _err := client.DescribeNetworkInterfacesWithOptions(DescribeNetworkInterfacesRequest, runtime)
	if _err != nil {
		errorHandler(_err)
		return VServerGroupBackendSpecs, fmt.Errorf("error DescribeNetworkInterfaces: %s", err)
	}
	// DescribeNetworkInterfaces.Body.NetworkInterfaceSets.NetworkInterfaceSet
	for _, networkInterface := range DescribeNetworkInterfaces.Body.NetworkInterfaceSets.NetworkInterfaceSet {
		for _, ip := range ServerIpsArray {
			if *networkInterface.PrivateIpAddress == *ip {
				VServerGroupBackendSpec := VServerGroupBackendSpec{}
				VServerGroupBackendSpec.Weight = Weight
				VServerGroupBackendSpec.Port = backendPort
				VServerGroupBackendSpec.ServerId = *networkInterface.InstanceId
				VServerGroupBackendSpec.ServerIp = *networkInterface.PrivateIpAddress
				VServerGroupBackendSpec.Type = Type
				VServerGroupBackendSpecs = append(VServerGroupBackendSpecs, VServerGroupBackendSpec)
			}
		}
	}
	return VServerGroupBackendSpecs, nil
}

func errorHandler(err error) {
	if err == nil {
		return
	}
	error := &tea.SDKError{}
	if _t, ok := err.(*tea.SDKError); ok {
		error = _t
	} else {
		error.Message = tea.String(err.Error())
	}
	Logger("error", fmt.Sprintf("errorHandler %s", tea.StringValue(error.Message)))
}
