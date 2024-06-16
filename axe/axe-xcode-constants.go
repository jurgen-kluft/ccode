package axe

type XcodeProductTypeEnum string

const (
	XcodeProductTypeApplication XcodeProductTypeEnum = "com.apple.product-type.application"
	XcodeProductTypeFramework   XcodeProductTypeEnum = "com.apple.product-type.framework"
	XcodeProductTypeStaticLib   XcodeProductTypeEnum = "com.apple.product-type.library.static"
	XcodeProductTypeDynamicLib  XcodeProductTypeEnum = "com.apple.product-type.library.dynamic"
	XcodeProductTypeTool        XcodeProductTypeEnum = "com.apple.product-type.tool"
)

func (p XcodeProductTypeEnum) String() string {
	return string(p)
}
