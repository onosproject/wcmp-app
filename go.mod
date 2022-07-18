module github.com/onosproject/wcmp-app

go 1.16

require (
	github.com/atomix/atomix-go-client v0.6.2
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.3.0
	github.com/onosproject/helmit v0.6.19
	github.com/onosproject/onos-api/go v0.9.26
	github.com/onosproject/onos-lib-go v0.8.16
	github.com/onosproject/onos-ric-sdk-go v0.8.9
	github.com/onosproject/onos-test v0.6.6
	github.com/onosproject/onos-topo v0.9.5
	github.com/p4lang/p4runtime v1.4.0-rc.5
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71
	google.golang.org/grpc v1.41.0
	k8s.io/client-go v0.22.1
)

replace github.com/onosproject/onos-api/go => /Users/arastega/go/src/github.com/onosproject/onos-api/go
