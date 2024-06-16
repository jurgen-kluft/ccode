package ide

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
