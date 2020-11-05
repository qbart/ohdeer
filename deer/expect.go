package deer

// Expect defines service assertion.
type Expect struct {
	// label
	Subject string `hcl:"subject,label"`
	// body
	Inclusion []int `hcl:"in"`
}
