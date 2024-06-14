package xcode

type ProductTypeEnum string

const (
	ProductTypeApplication ProductTypeEnum = "com.apple.product-type.application"
	ProductTypeFramework   ProductTypeEnum = "com.apple.product-type.framework"
	ProductTypeStaticLib   ProductTypeEnum = "com.apple.product-type.library.static"
	ProductTypeDynamicLib  ProductTypeEnum = "com.apple.product-type.library.dynamic"
	ProductTypeTool        ProductTypeEnum = "com.apple.product-type.tool"
)

func (p ProductTypeEnum) String() string {
	return string(p)
}

var BooleanToYesNo = map[bool]string{
	true:  "YES",
	false: "NO",
}
var BooleanToTrueFalse = map[bool]string{
	true:  "true",
	false: "false",
}

type Boolean bool

func (b Boolean) YesNo() string     { return BooleanToYesNo[bool(b)] }
func (b Boolean) TrueFalse() string { return BooleanToTrueFalse[bool(b)] }
func (b Boolean) Bool() bool        { return bool(b) }
func (b Boolean) Set(v bool)        { b = Boolean(v) }
func (b Boolean) Get() bool         { return bool(b) }
