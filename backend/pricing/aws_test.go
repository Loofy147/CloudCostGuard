package pricing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	t.Run("loads a valid pricing file", func(t *testing.T) {
		priceList := NewPriceList()
		err := priceList.LoadFromFile("../../testdata/sample-pricing.json")
		assert.NoError(t, err)

		// Check that the product was loaded
		sku := "JRTCKXETXF"
		product, ok := priceList.Products[sku]
		assert.True(t, ok, "Product with SKU %s should be loaded", sku)
		assert.Equal(t, "t2.micro", product.Attributes.InstanceType)

		// Check that the on-demand terms were loaded
		terms, ok := priceList.Terms.OnDemand[sku]
		assert.True(t, ok, "OnDemand terms for SKU %s should be loaded", sku)

		// Check a specific price dimension
		term, ok := terms[sku+".JRTCKXETXF"]
		assert.True(t, ok)

		priceDimension, ok := term.PriceDimensions[sku+".JRTCKXETXF.6YS6EN2CT7"]
		assert.True(t, ok)
		assert.Equal(t, "0.0116000000", priceDimension.PricePerUnit.USD)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		priceList := NewPriceList()
		err := priceList.LoadFromFile("non-existent-file.json")
		assert.Error(t, err)
	})
}
