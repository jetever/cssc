package transformer_test

import (
	"testing"

	"github.com/stephen/cssc/api/transforms"
	"github.com/stephen/cssc/internal/transformer"
	"github.com/stretchr/testify/assert"
)

func compileMediaQueryRanges(o *transformer.Options) {
	o.MediaFeatureRanges = transforms.MediaFeatureRangesTransform
}

func TestMediaQueryRanges(t *testing.T) {

	assert.Equal(t, `@media (min-width:200px) (max-width:600px),(min-width:200px),(max-width:600px){}`,
		Transform(t, compileMediaQueryRanges, `@media (200px <= width <= 600px), (200px <= width), (width <= 600px) {}`))

	assert.Equal(t, `@media (max-width:200px) (min-width:600px),(max-width:200px),(min-width:600px){}`,
		Transform(t, compileMediaQueryRanges, `@media (200px >= width >= 600px), (200px >= width), (width >= 600px) {}`))
}

func TestMediaQueryRanges_Passthrough(t *testing.T) {
	assert.Equal(t, `@media (200px>=width>=600px),(200px>=width),(width>=600px){}`,
		Transform(t, nil, `@media (200px >= width >= 600px), (200px >= width), (width >= 600px) {}`))
}

func TestMediaQueryRanges_Unsupported(t *testing.T) {
	assert.Panics(t, func() { Transform(t, compileMediaQueryRanges, `@media (200px < width < 600px) {}`) })
	assert.Panics(t, func() { Transform(t, compileMediaQueryRanges, `@media (width < 600px) {}`) })
	assert.Panics(t, func() { Transform(t, compileMediaQueryRanges, `@media (200px < width) {}`) })
	assert.Panics(t, func() { Transform(t, compileMediaQueryRanges, `@media (200px > width > 600px) {}`) })
	assert.Panics(t, func() { Transform(t, compileMediaQueryRanges, `@media (width > 600px) {}`) })
	assert.Panics(t, func() { Transform(t, compileMediaQueryRanges, `@media (200px > width) {}`) })
}
