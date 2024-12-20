package kubernetescontroller

func stringPtr(s string) *string {
	return &s
}

func Int32Ptr(s int32) *int32 {
	return &s
}

func TestAliyunclb() {
	aliyunclb := Aliyunclb{}
	// regionid := stringPtr("cn-hangzhou")
	aliyunclb.regionId = stringPtr("cn-hangzhou")
	aliyunclb.loadBalancerId = stringPtr("lb-bp117qz1mztx0r64oo6eg")
	aliyunclb.ListenerPort = Int32Ptr(80)
	aliyunclb.Protocol = stringPtr("UDP")
	aliyunclb.Description = stringPtr("test")
	aliyunclb.VServerGroupName = stringPtr("test")
	aliyunclb.VPCId = stringPtr("vpc-bp1km8lhqausyndk3gwol")
	ServerIps := []string{}
	ServerIps = append(ServerIps, "172.21.233.241")
	VServerGroupBackendSpecs, err := AliyunEcs(ServerIps, 100, 809, "cn-hangzhou", "ecs", *aliyunclb.VPCId)
	if err != nil {
		panic(err)
	}
	aliyunclb.VServerGroupBackendSpec = VServerGroupBackendSpecs

	aliyunclb._main() // init
	// err = aliyunclb.CreateSLBListener()
	// err = aliyunclb.DeleteSLBListener()
	errorHandler(err)
}
