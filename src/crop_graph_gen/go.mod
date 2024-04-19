module github.com/sannyschulz/crop-graph-gen/src/crop_graph_gen

go 1.21.5

require github.com/sannyschulz/crop-graph-gen/cropgraph v0.0.0-20240419113301-5368189bfc0c

replace github.com/sannyschulz/crop-graph-gen/cropgraph => ../../cropgraph

require (
	github.com/go-echarts/go-echarts/v2 v2.3.3 // indirect
	gopkg.in/yaml.v3 v3.0.0 // indirect
)
