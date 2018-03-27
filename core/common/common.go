package common

// constant for provider and consumer
const (
	Provider = "Provider"
	Consumer = "Consumer"
)

// constant for transport tcp
const (
	TransportTCP = "tcp"
)

// constant for microservice environment parameters
const (
	Env = "ServiceComb_ENV"

	EnvNodeIP     = "HOSTING_SERVER_IP"
	EnvSchemaRoot = "SCHEMA_ROOT"
	EnvProjectID  = "CSE_PROJECT_ID"
)

// constant environment keys service center, config center, monitor server addresses
const (
	CseRegistryAddress     = "CSE_REGISTRY_ADDR"
	CseConfigCenterAddress = "CSE_CONFIG_CENTER_ADDR"
	CseMonitorServer       = "CSE_MONITOR_SERVER_ADDR"
)

// env connect with "." like service_description.name and service_description.version which can not be used in k8s.
// So we can not use archaius to set env.
// To support this decalring constant for service name and version
// constant for service name and version.
const (
	ServiceName = "SERVICE_NAME"
	Version     = "VERSION"
)

// constant for microservice environment
const (
	EnvValueDev  = "development"
	EnvValueProd = "production"
)

// constant for secure socket layer parameters
const (
	SslCipherPluginKey = "cipherPlugin"
	SslVerifyPeerKey   = "verifyPeer"
	SslCipherSuitsKey  = "cipherSuits"
	SslProtocolKey     = "protocol"
	SslCaFileKey       = "caFile"
	SslCertFileKey     = "certFile"
	SslKeyFileKey      = "keyFile"
	SslCertPwdFileKey  = "certPwdFile"
	AKSKCustomCipher   = "cse.credentials.akskCustomCipher"
)

// constant for protocol types
const (
	ProtocolRest    = "rest"
	ProtocolHighway = "highway"
	LBSessionID     = "ServiceCombLB"
)

// DefaultKey default key
const DefaultKey = "default"

// DefaultValue default value
const DefaultValue = "default"

// BuildinTagApp build tag for the application
const BuildinTagApp = "app"

// BuildinTagVersion build tag version
const BuildinTagVersion = "version"

// CallerKey caller key
const CallerKey = "caller"

const (
	// HeaderSourceName is constant for header source name
	HeaderSourceName = "x-cse-src-microservice"
)

// constant for default application name and version
const (
	DefaultApp     = "default"
	DefaultVersion = "0.0.1"
	LatestVersion  = "latest"
	AllVersion     = "0+"
)

//constant used
const (
	HTTPS             = "https"
	JSON              = "application/json"
	Create            = "CREATE"
	Update            = "UPDATE"
	Delete            = "DELETE"
	Size              = "size"
	Client            = "client"
	File              = "File"
	SessionID         = "sessionid"
	ContentTypeJSON   = "application/json"
	DefaultTenant     = "default"
	DefaultChainName  = "default"
	RollingPolicySize = "size"
	FileRegistry      = "File"
	DefaultUserName   = "default"
	DefaultDomainName = "default"
	DefaultProvider   = "default"
)

// const default config for config-center
const (
	DefaultRefreshMode = 1
)
