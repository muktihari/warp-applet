package icon

import _ "embed"

//go:embed connected.png
var Connected []byte

//go:embed disconnected.png
var Disconnected []byte

//go:embed unknown_state.png
var UnknownState []byte
