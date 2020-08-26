package transformer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyLink(t *testing.T) {
	assert.Equal(t, ".test:visited,.test:link{color:red}", Transform(`
.test:any-link {
	color: red;
}`))

	assert.Equal(t, "complex .test:visited:not(.thing),complex .test:link:not(.thing){color:red}", Transform(`
complex .test:any-link:not(.thing) {
	color: red;
}`))

	assert.Equal(t, "a:visited,a:link,section,.Something{color:red}", Transform(`
a:any-link, section, .Something {
	color: red;
}`))
}