package constants

// These labels are only used on the ClusterIP services
// acting as each member's identity (static ip).
// Each of these labels is a record of intent to do
// something. The controller sets these labels and each
// member watches for them and takes the appropriate
// actions.
//
// See the sidecar design doc for more details.
const (
	// SeedLabel determines if a member is a seed or not.
	SeedLabel = "cassandra.rook.io/seed"

	// DecommissionLabel expresses the intent to decommission
	// the specific member. The presence of the label expresses
	// the intent to decommission. If the value is true, it means
	// the member has finished decommissioning.
	// Values: {true, false}
	DecommissionLabel = "cassandra.rook.io/decommissioned"

	// DeveloperModeAnnotation is present when the user wishes
	// to bypass production-readiness checks and start the database
	// either way. Currently useful for scylla, may get removed
	// once configMapName field is implemented in Cluster CRD.
	DeveloperModeAnnotation = "cassandra.rook.io/developer-mode"
	// CPUPinningAnnotation is present when the user has enabled
	// the CPU-Manager static policy and wants scylla to use
	// cpu pinning for improved performance.
	CPUPinningAnnotation = "cassandra.rook.io/cpu-pinning"

	LabelValueTrue  = "true"
	LabelValueFalse = "false"
)

// Generic Labels used on objects created by the operator.
const (
	ClusterNameLabel    = "cassandra.rook.io/cluster"
	DatacenterNameLabel = "cassandra.rook.io/datacenter"
	RackNameLabel       = "cassandra.rook.io/rack"

	AppName         = "rook-cassandra"
	OperatorAppName = "rook-cassandra-operator"
)

// Environment Variable Names
const (
	PodIPEnvVar = "POD_IP"

	ResourceLimitCPUEnvVar    = "CPU_LIMIT"
	ResourceLimitMemoryEnvVar = "MEMORY_LIMIT"
)

// Configuration Values
const (
	SharedDirName = "/mnt/shared"
	PluginDirName = SharedDirName + "/" + "plugins"

	DataDirCassandra = "/var/lib/cassandra/data"
	DataDirScylla    = "/var/lib/scylla/data"

	JolokiaJarName = "jolokia.jar"
	JolokiaPort    = 8778
	JolokiaContext = "jolokia"

	ReadinessProbePath = "/readyz"
	LivenessProbePath  = "/healthz"
	ProbePort          = 8080
)
