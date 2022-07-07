module github.com/onosproject/wcmp-app

go 1.16

require (
	github.com/google/uuid v1.1.2
	github.com/onosproject/onos-api/go v0.0.0-00010101000000-000000000000
	github.com/onosproject/onos-lib-go v0.8.16
	github.com/p4lang/p4runtime v1.3.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71
	google.golang.org/grpc v1.41.0
)

replace github.com/onosproject/onos-api/go => github.com/adibrastegarnia/onos-api/go v0.9.9-0.20220707213709-76695c383cfc
