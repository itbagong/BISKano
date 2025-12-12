package scmlogic_test

import (
	"testing"

	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestInventDim(t *testing.T) {
	convey.Convey("create dim1", t, func() {
		dim1 := scmmodel.InventDimension{VariantID: "Type1", Grade: "A", WarehouseID: "HO"}
		dim1.Calc()
		convey.SoMsg("SpecID != blank", dim1.SpecID, convey.ShouldNotBeBlank)
		convey.SoMsg("InventDim != blank", dim1.InventDimID, convey.ShouldNotBeBlank)

		convey.Convey("create dim2 with same spec different section", func() {
			dim2 := scmmodel.InventDimension{VariantID: "Type1", Grade: "A", WarehouseID: "HO", SectionID: "SCM"}
			dim2.Calc()
			convey.SoMsg("SpecID != dim1.SpecID", dim2.SpecID, convey.ShouldEqual, dim1.SpecID)
			convey.SoMsg("InventDim != blank", dim2.InventDimID, convey.ShouldNotBeBlank)
			convey.SoMsg("InventDim != dim1.InventDimID", dim2.InventDimID, convey.ShouldNotEqual, dim1.InventDimID)

			convey.Println()
			convey.Println("dim1:", codekit.JsonString(dim1))
			convey.Println("dim2:", codekit.JsonString(dim2))
		})
	})
}
