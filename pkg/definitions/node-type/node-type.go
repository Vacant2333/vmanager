package node_type

type NodeType string

const (
	NodeTypeLabelKey = "node.kubernetes.io/capacity"

	NodeTypeOnDemand NodeType = "on-demand"
	NodeTypeSpot     NodeType = "spot"
)
