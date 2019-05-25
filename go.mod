module github.com/bryanl/keps

go 1.12

require (
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/colour v0.0.0-20160524082231-60882d9e2721 // indirect
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/ghodss/yaml v1.0.0
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	gopkg.in/russross/blackfriday.v2 v2.0.1
	gopkg.in/yaml.v2 v2.2.2
)

replace gopkg.in/russross/blackfriday.v2 => github.com/russross/blackfriday/v2 v2.0.1
