package xchg

import "strings"

const (
	ERR_XCHG_ACCESS_DENIED = "{ERR_XCHG_ACCESS_DENIED}"

	// Base Connection
	ERR_XCHG_CONN_WRONG_FRAME_SIZE = "{ERR_XCHG_CONN_WRONG_FRAME_SIZE}"
	ERR_XCHG_CONN_NO_CONNECTION    = "{ERR_XCHG_CONN_NO_CONNECTION}"
	ERR_XCHG_CONN_SENDING_ERROR    = "{ERR_XCHG_CONN_SENDING_ERROR}"

	// Transaction
	ERR_XCHG_TR_WRONG_FRAME = "{ERR_XCHG_TR_WRONG_FRAME}"

	// Client Connection
	// Regular Call
	ERR_XCHG_CL_CONN_CALL_WRONG_FUNCTION_LEN   = "{ERR_XCHG_CL_CONN_CALL_WRONG_FUNCTION_LEN}"
	ERR_XCHG_CL_CONN_CALL_SEARCHING_NODE       = "{ERR_XCHG_CL_CONN_CALL_SEARCHING_NODE}"
	ERR_XCHG_CL_CONN_CALL_NO_LOCAL_PRIVATE_KEY = "{ERR_XCHG_CL_CONN_CALL_NO_LOCAL_PRIVATE_KEY}"
	ERR_XCHG_CL_CONN_CALL_NO_ROUTE_TO_PEER     = "{ERR_XCHG_CL_CONN_CALL_NO_ROUTE_TO_PEER}"
	ERR_XCHG_CL_CONN_CALL_RESP_LEN             = "{ERR_XCHG_CL_CONN_CALL_RESP_LEN}"
	ERR_XCHG_CL_CONN_CALL_RESP_STATUS_BYTE     = "{ERR_XCHG_CL_CONN_CALL_RESP_STATUS_BYTE}"
	ERR_XCHG_CL_CONN_CALL_ENC                  = "{ERR_XCHG_CL_CONN_CALL_ENC}"
	ERR_XCHG_CL_CONN_CALL_ERR                  = "{ERR_XCHG_CL_CONN_CALL_ERR}"
	ERR_XCHG_CL_CONN_CALL_FROM_PEER            = "{ERR_XCHG_CL_CONN_FROM_PEER}"

	// Auth
	ERR_XCHG_CL_CONN_AUTH_GET_NONCE            = "{ERR_XCHG_CL_CONN_AUTH_GET_NONCE}"
	ERR_XCHG_CL_CONN_AUTH_WRONG_NONCE_LEN      = "{ERR_XCHG_CL_CONN_AUTH_WRONG_NONCE_LEN}"
	ERR_XCHG_CL_CONN_AUTH_NO_LOCAL_PRIVATE_KEY = "{ERR_XCHG_CL_CONN_AUTH_NO_LOCAL_PRIVATE_KEY}"
	ERR_XCHG_CL_CONN_AUTH_ENC                  = "{ERR_XCHG_CL_CONN_AUTH_ENC}"
	ERR_XCHG_CL_CONN_AUTH_AUTH                 = "{ERR_XCHG_CL_CONN_AUTH_AUTH}"
	ERR_XCHG_CL_CONN_AUTH_DECR                 = "{ERR_XCHG_CL_CONN_AUTH_DECR}"
	ERR_XCHG_CL_CONN_AUTH_WRONG_AUTH_RESP_LEN  = "{ERR_XCHG_CL_CONN_AUTH_WRONG_AUTH_RESP_LEN}"

	// Peer Connection
	ERR_XCHG_PEER_CONN_LOSS               = "{ERR_XCHG_PEER_CONN_LOSS}"
	ERR_XCHG_PEER_CONN_TR_TIMEOUT         = "{ERR_XCHG_PEER_CONN_TR_TIMEOUT}"
	ERR_XCHG_PEER_CONN_REQ_SID_SIZE       = "{ERR_XCHG_PEER_CONN_REQ_SID_SIZE}"
	ERR_XCHG_PEER_CONN_WRONG_PROT_VERSION = "{ERR_XCHG_PEER_CONN_WRONG_PROT_VERSION}"
	ERR_XCHG_PEER_CONN_RCVD_ERR           = "{ERR_XCHG_PEER_CONN_RCVD_ERR}"

	// Server Connection
	ERR_XCHG_SRV_CONN_WRONG_SESSION       = "{ERR_XCHG_SRV_CONN_WRONG_SESSION}"
	ERR_XCHG_SRV_CONN_DECR                = "{ERR_XCHG_SRV_CONN_DECR}"
	ERR_XCHG_SRV_CONN_WRONG_LEN9          = "{ERR_XCHG_SRV_CONN_WRONG_LEN9}"
	ERR_XCHG_SRV_CONN_WRONG_NONCE         = "{ERR_XCHG_SRV_CONN_WRONG_NONCE}"
	ERR_XCHG_SRV_CONN_WRONG_LEN1          = "{ERR_XCHG_SRV_CONN_WRONG_LEN1}"
	ERR_XCHG_SRV_CONN_WRONG_LEN_FN        = "{ERR_XCHG_SRV_CONN_WRONG_LEN_FN}"
	ERR_XCHG_SRV_CONN_AUTH_DATA_LEN4      = "{ERR_XCHG_SRV_CONN_AUTH_DATA_LEN4}"
	ERR_XCHG_SRV_CONN_NOT_IMPL            = "{ERR_XCHG_SRV_CONN_NOT_IMPL}"
	ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_NONCE = "{ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_NONCE}"
	ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_PK    = "{ERR_XCHG_SRV_CONN_AUTH_DATA_LEN_PK}"
	ERR_XCHG_SRV_CONN_AUTH_WRONG_NONCE    = "{ERR_XCHG_SRV_CONN_AUTH_WRONG_NONCE}"

	// Router
	ERR_XCHG_ROUTER_CONFIG_IS_DIRECTORY         = "{ERR_XCHG_ROUTER_CONFIG_IS_DIRECTORY}"
	ERR_XCHG_ROUTER_CONN_WRONG_FRAME_TYPE       = "{ERR_XCHG_ROUTER_CONN_WRONG_FRAME_TYPE}"
	ERR_XCHG_ROUTER_CONN_WRONG_PROTOCOL_VERSION = "{ERR_XCHG_ROUTER_CONN_WRONG_PROTOCOL_VERSION}"
	ERR_XCHG_ROUTER_CONN_WRONG_PUBLIC_KEY_SIZE  = "{ERR_XCHG_ROUTER_CONN_WRONG_PUBLIC_KEY_SIZE}"
	ERR_XCHG_ROUTER_CONN_ENC                    = "{ERR_XCHG_ROUTER_CONN_ENC}"
	ERR_XCHG_ROUTER_CONN_DECR4                  = "{ERR_XCHG_ROUTER_CONN_DECR4}"
	ERR_XCHG_ROUTER_CONN_DECR5                  = "{ERR_XCHG_ROUTER_CONN_DECR5}"
	ERR_XCHG_ROUTER_CONN_NO_ROUTE_TO_PEER       = "{ERR_XCHG_ROUTER_CONN_NO_ROUTE_TO_PEER}"
	ERR_XCHG_ROUTER_SERVER_ALREADY_STARTED      = "{ERR_XCHG_ROUTER_SERVER_ALREADY_STARTED}"
	ERR_XCHG_ROUTER_SERVER_IS_NOT_STARTED       = "{ERR_XCHG_ROUTER_SERVER_IS_NOT_STARTED}"
	ERR_XCHG_ROUTER_ALREADY_STARTED             = "{ERR_XCHG_ROUTER_ALREADY_STARTED}"
	ERR_XCHG_ROUTER_IS_NOT_STARTED              = "{ERR_XCHG_ROUTER_IS_NOT_STARTED}"

	// Other
	ERR_SIMPLE_SERVER_FUNC_IS_NOT_IMPL = "{ERR_SIMPLE_SERVER_FUNC_IS_NOT_IMPL}"
)

// Reason to make session
func NeedToMakeSession(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, "{ERR_XCHG_SRV_CONN_") {
		return true
	}
	return false
}

// Reason to change node
func NeedToChangeNode(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	if strings.Contains(errStr, "{ERR_XCHG_ROUTER_") {
		return true
	}
	return false
}
