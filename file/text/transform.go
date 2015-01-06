package text

import (

)

type Transformer interface {
	Transform(src string) (dst string, err error)
}
