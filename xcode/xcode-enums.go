package xcode

type ProductTypeEnum string

const (
	ProductTypeApplication ProductTypeEnum = "com.apple.product-type.application"
	ProductTypeFramework   ProductTypeEnum = "com.apple.product-type.framework"
	ProductTypeStaticLib   ProductTypeEnum = "com.apple.product-type.library.static"
)
