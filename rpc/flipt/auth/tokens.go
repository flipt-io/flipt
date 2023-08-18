package auth

// Strips the client token from the CallbackResponse.
func (cbr *CallbackResponse) StripClientToken() {
	cbr.ClientToken = ""
}

// Strips the client token from the OAuthCallbackResponse.
func (ocbr *OAuthCallbackResponse) StripClientToken() {
	ocbr.ClientToken = ""
}
