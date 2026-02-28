package enums

// ReroutingDisconnectCode defines SIP disconnect codes that trigger rerouting for voice-in trunks.
type ReroutingDisconnectCode int

const (
	DCSIP400BadRequest              ReroutingDisconnectCode = 56
	DCSIP401Unauthorized            ReroutingDisconnectCode = 57
	DCSIP402PaymentRequired         ReroutingDisconnectCode = 58
	DCSIP403Forbidden               ReroutingDisconnectCode = 59
	DCSIP404NotFound                ReroutingDisconnectCode = 60
	DCSIP408RequestTimeout          ReroutingDisconnectCode = 64
	DCSIP409Conflict                ReroutingDisconnectCode = 65
	DCSIP410Gone                    ReroutingDisconnectCode = 66
	DCSIP412ConditionalRequestFail  ReroutingDisconnectCode = 67
	DCSIP413RequestEntityTooLarge   ReroutingDisconnectCode = 68
	DCSIP414RequestURITooLong       ReroutingDisconnectCode = 69
	DCSIP415UnsupportedMediaType    ReroutingDisconnectCode = 70
	DCSIP416UnsupportedURIScheme    ReroutingDisconnectCode = 71
	DCSIP417UnknownResourcePriority ReroutingDisconnectCode = 72
	DCSIP420BadExtension            ReroutingDisconnectCode = 73
	DCSIP421ExtensionRequired       ReroutingDisconnectCode = 74
	DCSIP422SessionIntervalTooSmall ReroutingDisconnectCode = 75
	DCSIP423IntervalTooBrief        ReroutingDisconnectCode = 76
	DCSIP424BadLocationInformation  ReroutingDisconnectCode = 77
	DCSIP428UseIdentityHeader       ReroutingDisconnectCode = 78
	DCSIP429ProvideReferrerIdentity ReroutingDisconnectCode = 79
	DCSIP433AnonymityDisallowed     ReroutingDisconnectCode = 80
	DCSIP436BadIdentityInfo         ReroutingDisconnectCode = 81
	DCSIP437UnsupportedCertificate  ReroutingDisconnectCode = 82
	DCSIP438InvalidIdentityHeader   ReroutingDisconnectCode = 83
	DCSIP480TemporarilyUnavailable  ReroutingDisconnectCode = 84
	DCSIP482LoopDetected            ReroutingDisconnectCode = 86
	DCSIP483TooManyHops             ReroutingDisconnectCode = 87
	DCSIP484AddressIncomplete       ReroutingDisconnectCode = 88
	DCSIP485Ambiguous               ReroutingDisconnectCode = 89
	DCSIP486BusyHere                ReroutingDisconnectCode = 90
	DCSIP487RequestTerminated       ReroutingDisconnectCode = 91
	DCSIP488NotAcceptableHere       ReroutingDisconnectCode = 92
	DCSIP494SecurityAgreementReq    ReroutingDisconnectCode = 96
	DCSIP500ServerInternalError     ReroutingDisconnectCode = 97
	DCSIP501NotImplemented          ReroutingDisconnectCode = 98
	DCSIP502BadGateway              ReroutingDisconnectCode = 99
	DCSIP503ServiceUnavailable      ReroutingDisconnectCode = 100
	DCSIP504ServerTimeout           ReroutingDisconnectCode = 101
	DCSIP505VersionNotSupported     ReroutingDisconnectCode = 102
	DCSIP513MessageTooLarge         ReroutingDisconnectCode = 103
	DCSIP580PreconditionFailure     ReroutingDisconnectCode = 104
	DCSIP600BusyEverywhere          ReroutingDisconnectCode = 105
	DCSIP603Decline                 ReroutingDisconnectCode = 106
	DCSIP604DoesNotExistAnywhere    ReroutingDisconnectCode = 107
	DCSIP606NotAcceptable           ReroutingDisconnectCode = 108
	DCRingingTimeout                ReroutingDisconnectCode = 1505
)
