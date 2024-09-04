package common

import (
	"math"
	"time"

	"github.com/DharitriOne/drt-chain-core-go/core"
)

// NodeOperation defines the p2p node operation
type NodeOperation string

// NormalOperation defines the normal mode operation: either seeder, observer or validator
const NormalOperation NodeOperation = "normal operation"

// FullArchiveMode defines the node operation as a full archive mode
const FullArchiveMode NodeOperation = "full archive mode"

// PeerType represents the type of peer
type PeerType string

// EligibleList represents the list of peers who participate in consensus inside a shard
const EligibleList PeerType = "eligible"

// WaitingList represents the list of peers who don't participate in consensus but will join the next epoch
const WaitingList PeerType = "waiting"

// LeavingList represents the list of peers who were taken out of eligible and waiting because of rating
const LeavingList PeerType = "leaving"

// InactiveList represents the list of peers who were taken out because they were leaving
const InactiveList PeerType = "inactive"

// JailedList represents the list of peers who have stake but are in jail
const JailedList PeerType = "jailed"

// ObserverList represents the list of peers who don't participate in consensus but will join the next epoch
const ObserverList PeerType = "observer"

// NewList represents the list of peers who have stake and are pending to become eligible
const NewList PeerType = "new"

// MetachainTopicIdentifier is the identifier used in topics to define the metachain shard ID
const MetachainTopicIdentifier = "META" // TODO - move this to drt-chain-core-go and change wherever we use the string value

// AuctionList represents the list of peers which don't participate in consensus yet, but will be selected
// based on their top up stake
const AuctionList PeerType = "auction"

// SelectedFromAuctionList represents the list of peers which have been selected from AuctionList based on
// their top up to be distributed on the WaitingList in the next epoch
const SelectedFromAuctionList PeerType = "selectedFromAuction"

// CombinedPeerType - represents the combination of two peerTypes
const CombinedPeerType = "%s (%s)"

// UnVersionedAppString represents the default app version that indicate that the binary wasn't build by setting
// the appVersion flag
const UnVersionedAppString = "undefined"

// DisabledShardIDAsObserver defines the uint32 identifier which tells that the node hasn't configured any preferred
// shard to start in as observer
const DisabledShardIDAsObserver = uint32(0xFFFFFFFF) - 7

// MaxTxNonceDeltaAllowed specifies the maximum difference between an account's nonce and a received transaction's nonce
// in order to mark the transaction as valid.
const MaxTxNonceDeltaAllowed = 100

// MaxBulkTransactionSize specifies the maximum size of one bulk with txs which can be send over the network
// TODO convert this const into a var and read it from config when this code moves to another binary
const MaxBulkTransactionSize = 1 << 18 // 256KB bulks

// MaxTxsToRequest specifies the maximum number of txs to request
const MaxTxsToRequest = 1000

// NodesSetupJsonFileName specifies the name of the json file which contains the setup of the nodes
const NodesSetupJsonFileName = "nodesSetup.json"

// ConsensusTopic is the topic used in consensus algorithm
const ConsensusTopic = "consensus"

// GenesisTxSignatureString is the string used to generate genesis transaction signature as 128 hex characters
const GenesisTxSignatureString = "GENESISGENESISGENESISGENESISGENESISGENESISGENESISGENESISGENESISG"

// HeartbeatV2Topic is the topic used for heartbeatV2 signaling
const HeartbeatV2Topic = "heartbeatV2"

// PeerAuthenticationTopic is the topic used for peer authentication signaling
const PeerAuthenticationTopic = "peerAuthentication"

// ConnectionTopic represents the topic used when sending the new connection message data
const ConnectionTopic = "connection"

// ValidatorInfoTopic is the topic used for validatorInfo signaling
const ValidatorInfoTopic = "validatorInfo"

// MetricCurrentRound is the metric for monitoring the current round of a node
const MetricCurrentRound = "moa_current_round"

// MetricNonce is the metric for monitoring the nonce of a node
const MetricNonce = "moa_nonce"

// MetricBlockTimestamp is the metric for monitoring the timestamp of the last synchronized block
const MetricBlockTimestamp = "moa_block_timestamp"

// MetricProbableHighestNonce is the metric for monitoring the max speculative nonce received by the node by listening on the network
const MetricProbableHighestNonce = "moa_probable_highest_nonce"

// MetricNumConnectedPeers is the metric for monitoring the number of connected peers
const MetricNumConnectedPeers = "moa_num_connected_peers"

// MetricNumConnectedPeersClassification is the metric for monitoring the number of connected peers split on the connection type
const MetricNumConnectedPeersClassification = "moa_num_connected_peers_classification"

// MetricSynchronizedRound is the metric for monitoring the synchronized round of a node
const MetricSynchronizedRound = "moa_synchronized_round"

// MetricIsSyncing is the metric for monitoring if a node is syncing
const MetricIsSyncing = "moa_is_syncing"

// MetricPublicKeyBlockSign is the metric for monitoring public key of a node used in block signing
const MetricPublicKeyBlockSign = "moa_public_key_block_sign"

// MetricShardId is the metric for monitoring shard id of a node
const MetricShardId = "moa_shard_id"

// MetricNumShardsWithoutMetachain is the metric for monitoring the number of shards (excluding meta)
const MetricNumShardsWithoutMetachain = "moa_num_shards_without_meta"

// MetricTxPoolLoad is the metric for monitoring number of transactions from pool of a node
const MetricTxPoolLoad = "moa_tx_pool_load"

// MetricCountLeader is the metric for monitoring number of rounds when a node was leader
const MetricCountLeader = "moa_count_leader"

// MetricCountConsensus is the metric for monitoring number of rounds when a node was in consensus group
const MetricCountConsensus = "moa_count_consensus"

// MetricCountAcceptedBlocks is the metric for monitoring number of blocks that was accepted proposed by a node
const MetricCountAcceptedBlocks = "moa_count_accepted_blocks"

// MetricNodeType is the metric for monitoring the type of the node
const MetricNodeType = "moa_node_type"

// MetricLiveValidatorNodes is the metric for the number of live validators on the network
const MetricLiveValidatorNodes = "moa_live_validator_nodes"

// MetricConnectedNodes is the metric for monitoring total connected nodes on the network
const MetricConnectedNodes = "moa_connected_nodes"

// MetricNumIntraShardValidatorNodes is the metric for the number of intra-shard validators
const MetricNumIntraShardValidatorNodes = "moa_intra_shard_validator_nodes"

// MetricCpuLoadPercent is the metric for monitoring CPU load [%]
const MetricCpuLoadPercent = "moa_cpu_load_percent"

// MetricMemLoadPercent is the metric for monitoring memory load [%]
const MetricMemLoadPercent = "moa_mem_load_percent"

// MetricMemTotal is the metric for monitoring total memory bytes
const MetricMemTotal = "moa_mem_total"

// MetricMemUsedGolang is a metric for monitoring the memory ("total")
const MetricMemUsedGolang = "moa_mem_used_golang"

// MetricMemUsedSystem is a metric for monitoring the memory ("sys mem")
const MetricMemUsedSystem = "moa_mem_used_sys"

// MetricMemHeapInUse is a metric for monitoring the memory ("heap in use")
const MetricMemHeapInUse = "moa_mem_heap_inuse"

// MetricMemStackInUse is a metric for monitoring the memory ("stack in use")
const MetricMemStackInUse = "moa_mem_stack_inuse"

// MetricNetworkRecvPercent is the metric for monitoring network receive load [%]
const MetricNetworkRecvPercent = "moa_network_recv_percent"

// MetricNetworkRecvBps is the metric for monitoring network received bytes per second
const MetricNetworkRecvBps = "moa_network_recv_bps"

// MetricNetworkRecvBpsPeak is the metric for monitoring network received peak bytes per second
const MetricNetworkRecvBpsPeak = "moa_network_recv_bps_peak"

// MetricNetworkRecvBytesInCurrentEpochPerHost is the metric for monitoring network received bytes in current epoch per host
const MetricNetworkRecvBytesInCurrentEpochPerHost = "moa_network_recv_bytes_in_epoch_per_host"

// MetricNetworkSendBytesInCurrentEpochPerHost is the metric for monitoring network send bytes in current epoch per host
const MetricNetworkSendBytesInCurrentEpochPerHost = "moa_network_sent_bytes_in_epoch_per_host"

// MetricNetworkSentPercent is the metric for monitoring network sent load [%]
const MetricNetworkSentPercent = "moa_network_sent_percent"

// MetricNetworkSentBps is the metric for monitoring network sent bytes per second
const MetricNetworkSentBps = "moa_network_sent_bps"

// MetricNetworkSentBpsPeak is the metric for monitoring network sent peak bytes per second
const MetricNetworkSentBpsPeak = "moa_network_sent_bps_peak"

// MetricRoundTime is the metric for round time in seconds
const MetricRoundTime = "moa_round_time"

// MetricEpochNumber is the metric for the number of epoch
const MetricEpochNumber = "moa_epoch_number"

// MetricAppVersion is the metric for the current app version
const MetricAppVersion = "moa_app_version"

// MetricNumTxInBlock is the metric for the number of transactions in the proposed block
const MetricNumTxInBlock = "moa_num_tx_block"

// MetricConsensusState is the metric for consensus state of node proposer,participant or not consensus group
const MetricConsensusState = "moa_consensus_state"

// MetricNumMiniBlocks is the metric for number of miniblocks in a block
const MetricNumMiniBlocks = "moa_num_mini_blocks"

// MetricConsensusRoundState is the metric for consensus round state for a block
const MetricConsensusRoundState = "moa_consensus_round_state"

// MetricCrossCheckBlockHeight is the metric that store cross block height
const MetricCrossCheckBlockHeight = "moa_cross_check_block_height"

// MetricCrossCheckBlockHeightMeta is the metric that store metachain cross block height
const MetricCrossCheckBlockHeightMeta = "moa_cross_check_block_height_meta"

// MetricNumProcessedTxs is the metric that stores the number of transactions processed
const MetricNumProcessedTxs = "moa_num_transactions_processed"

// MetricCurrentBlockHash is the metric that stores the current block hash
const MetricCurrentBlockHash = "moa_current_block_hash"

// MetricCurrentRoundTimestamp is the metric that stores current round timestamp
const MetricCurrentRoundTimestamp = "moa_current_round_timestamp"

// MetricHeaderSize is the metric that stores the current block size
const MetricHeaderSize = "moa_current_block_size"

// MetricMiniBlocksSize is the metric that stores the current block size
const MetricMiniBlocksSize = "moa_mini_blocks_size"

// MetricNumShardHeadersFromPool is the metric that stores number of shard header from pool
const MetricNumShardHeadersFromPool = "moa_num_shard_headers_from_pool"

// MetricNumShardHeadersProcessed is the metric that stores number of shard header processed
const MetricNumShardHeadersProcessed = "moa_num_shard_headers_processed"

// MetricNumTimesInForkChoice is the metric that counts how many times a node was in fork choice
const MetricNumTimesInForkChoice = "moa_fork_choice_count"

// MetricHighestFinalBlock is the metric for the nonce of the highest final block
const MetricHighestFinalBlock = "moa_highest_final_nonce"

// MetricLatestTagSoftwareVersion is the metric that stores the latest tag software version
const MetricLatestTagSoftwareVersion = "moa_latest_tag_software_version"

// MetricCountConsensusAcceptedBlocks is the metric for monitoring number of blocks accepted when the node was in consensus group
const MetricCountConsensusAcceptedBlocks = "moa_count_consensus_accepted_blocks"

// MetricNodeDisplayName is the metric that stores the name of the node
const MetricNodeDisplayName = "moa_node_display_name"

// MetricConsensusGroupSize is the metric for consensus group size for the current shard/meta
const MetricConsensusGroupSize = "moa_consensus_group_size"

// MetricShardConsensusGroupSize is the metric for the shard consensus group size
const MetricShardConsensusGroupSize = "moa_shard_consensus_group_size"

// MetricMetaConsensusGroupSize is the metric for the metachain consensus group size
const MetricMetaConsensusGroupSize = "moa_meta_consensus_group_size"

// MetricNumNodesPerShard is the metric which holds the number of nodes in a shard
const MetricNumNodesPerShard = "moa_num_nodes_in_shard"

// MetricNumMetachainNodes is the metric which holds the number of nodes in metachain
const MetricNumMetachainNodes = "moa_num_metachain_nodes"

// MetricNumValidators is the metric for the number of validators
const MetricNumValidators = "moa_num_validators"

// MetricPeerType is the metric which tells the peer's type (in eligible list, in waiting list, or observer)
const MetricPeerType = "moa_peer_type"

// MetricPeerSubType is the metric which tells the peer's subtype (regular observer or full history observer)
const MetricPeerSubType = "moa_peer_subtype"

// MetricLeaderPercentage is the metric for leader rewards percentage
const MetricLeaderPercentage = "moa_leader_percentage"

// MetricDenomination is the metric for exposing the denomination
const MetricDenomination = "moa_denomination"

// MetricRoundAtEpochStart is the metric for storing the first round of the current epoch
const MetricRoundAtEpochStart = "moa_round_at_epoch_start"

// MetricNonceAtEpochStart is the metric for storing the first nonce of the current epoch
const MetricNonceAtEpochStart = "moa_nonce_at_epoch_start"

// MetricRoundsPerEpoch is the metric that tells the number of rounds in an epoch
const MetricRoundsPerEpoch = "moa_rounds_per_epoch"

// MetricRoundsPassedInCurrentEpoch is the metric that tells the number of rounds passed in current epoch
const MetricRoundsPassedInCurrentEpoch = "moa_rounds_passed_in_current_epoch"

// MetricNoncesPassedInCurrentEpoch is the metric that tells the number of nonces passed in current epoch
const MetricNoncesPassedInCurrentEpoch = "moa_nonces_passed_in_current_epoch"

// MetricReceivedProposedBlock is the metric that specifies the moment in the round when the received block has reached the
// current node. The value is provided in percent (0 meaning it has been received just after the round started and
// 100 meaning that the block has been received in the last moment of the round)
const MetricReceivedProposedBlock = "moa_consensus_received_proposed_block"

// MetricCreatedProposedBlock is the metric that specifies the percent of the block subround used for header and body
// creation (0 meaning that the block was created in no-time and 100 meaning that the block creation used all the
// subround spare duration)
const MetricCreatedProposedBlock = "moa_consensus_created_proposed_block"

// MetricRedundancyLevel is the metric that specifies the redundancy level of the current node
const MetricRedundancyLevel = "moa_redundancy_level"

// MetricRedundancyIsMainActive is the metric that specifies data about the redundancy main machine
const MetricRedundancyIsMainActive = "moa_redundancy_is_main_active"

// MetricRedundancyStepInReason is the metric that specifies why the back-up machine stepped in
const MetricRedundancyStepInReason = "moa_redundancy_step_in_reason"

// MetricValueNA represents the value to be used when a metric is not available/applicable
const MetricValueNA = "N/A"

// MetricProcessedProposedBlock is the metric that specify the percent of the block subround used for header and body
// processing (0 meaning that the block was processed in no-time and 100 meaning that the block processing used all the
// subround spare duration)
const MetricProcessedProposedBlock = "moa_consensus_processed_proposed_block"

// MetricMinGasPrice is the metric that specifies min gas price
const MetricMinGasPrice = "moa_min_gas_price"

// MetricMinGasLimit is the metric that specifies the minimum gas limit
const MetricMinGasLimit = "moa_min_gas_limit"

// MetricExtraGasLimitGuardedTx specifies the extra gas limit required for guarded transactions
const MetricExtraGasLimitGuardedTx = "moa_extra_gas_limit_guarded_tx"

// MetricRewardsTopUpGradientPoint is the metric that specifies the rewards top up gradient point
const MetricRewardsTopUpGradientPoint = "moa_rewards_top_up_gradient_point"

// MetricGasPriceModifier is the metric that specifies the gas price modifier
const MetricGasPriceModifier = "moa_gas_price_modifier"

// MetricTopUpFactor is the metric that specifies the top-up factor
const MetricTopUpFactor = "moa_top_up_factor"

// MetricMinTransactionVersion is the metric that specifies the minimum transaction version
const MetricMinTransactionVersion = "moa_min_transaction_version"

// MetricGatewayMetricsEndpoint is the metric that specifies gateway endpoint
const MetricGatewayMetricsEndpoint = "moa_gateway_metrics_endpoint"

// MetricGasPerDataByte is the metric that specifies the required gas for a data byte
const MetricGasPerDataByte = "moa_gas_per_data_byte"

// MetricMaxGasPerTransaction is the metric that specifies the maximum gas limit for a transaction
const MetricMaxGasPerTransaction = "moa_max_gas_per_transaction"

// MetricChainId is the metric that specifies current chain id
const MetricChainId = "moa_chain_id"

// MetricStartTime is the metric that specifies the genesis start time
const MetricStartTime = "moa_start_time"

// MetricRoundDuration is the metric that specifies the round duration in milliseconds
const MetricRoundDuration = "moa_round_duration"

// MetricTotalSupply holds the total supply value for the last epoch
const MetricTotalSupply = "moa_total_supply"

// MetricTotalBaseStakedValue holds the total base staked value
const MetricTotalBaseStakedValue = "moa_total_base_staked_value"

// MetricTopUpValue holds the total top up value
const MetricTopUpValue = "moa_total_top_up_value"

// MetricInflation holds the inflation value for the last epoch
const MetricInflation = "moa_inflation"

// MetricDevRewardsInEpoch holds the developers' rewards value for the last epoch
const MetricDevRewardsInEpoch = "moa_dev_rewards"

// MetricTotalFees holds the total fees value for the last epoch
const MetricTotalFees = "moa_total_fees"

// MetricEpochForEconomicsData holds the epoch for which economics data are computed
const MetricEpochForEconomicsData = "moa_epoch_for_economics_data"

// MetachainShardId will be used to identify a shard ID as metachain
const MetachainShardId = uint32(0xFFFFFFFF)

// BaseOperationCost represents the field name for base operation costs
const BaseOperationCost = "BaseOperationCost"

// BuiltInCost represents the field name for built-in operation costs
const BuiltInCost = "BuiltInCost"

// MetaChainSystemSCsCost represents the field name for metachain system smart contract operation costs
const MetaChainSystemSCsCost = "MetaChainSystemSCsCost"

// BaseOpsAPICost represents the field name of the SC API (EEI) gas costs
const BaseOpsAPICost = "BaseOpsAPICost"

// MaxPerTransaction represents the field name of max counts per transaction in block chain hook
const MaxPerTransaction = "MaxPerTransaction"

// AsyncCallStepField is the field name for the gas cost for any of the two steps required to execute an async call
const AsyncCallStepField = "AsyncCallStep"

// AsyncCallbackGasLockField is the field name for the gas amount to be locked
// before executing the destination async call, to be put aside for the async callback
const AsyncCallbackGasLockField = "AsyncCallbackGasLock"

const (
	// MetricScDeployEnableEpoch represents the epoch when the deployment of smart contracts is enabled
	MetricScDeployEnableEpoch = "moa_smart_contract_deploy_enable_epoch"

	// MetricBuiltInFunctionsEnableEpoch represents the epoch when the built-in functions is enabled
	MetricBuiltInFunctionsEnableEpoch = "moa_built_in_functions_enable_epoch"

	// MetricRelayedTransactionsEnableEpoch represents the epoch when the relayed transactions is enabled
	MetricRelayedTransactionsEnableEpoch = "moa_relayed_transactions_enable_epoch"

	// MetricPenalizedTooMuchGasEnableEpoch represents the epoch when the penalization for using too much gas is enabled
	MetricPenalizedTooMuchGasEnableEpoch = "moa_penalized_too_much_gas_enable_epoch"

	// MetricSwitchJailWaitingEnableEpoch represents the epoch when the system smart contract processing at end of epoch is enabled
	MetricSwitchJailWaitingEnableEpoch = "moa_switch_jail_waiting_enable_epoch"

	// MetricSwitchHysteresisForMinNodesEnableEpoch represents the epoch when the system smart contract changes its config to consider
	// also (minimum) hysteresis nodes for the minimum number of nodes
	MetricSwitchHysteresisForMinNodesEnableEpoch = "moa_switch_hysteresis_for_min_nodes_enable_epoch"

	// MetricBelowSignedThresholdEnableEpoch represents the epoch when the change for computing rating for validators below signed rating is enabled
	MetricBelowSignedThresholdEnableEpoch = "moa_below_signed_threshold_enable_epoch"

	// MetricTransactionSignedWithTxHashEnableEpoch represents the epoch when the node will also accept transactions that are
	// signed with the hash of transaction
	MetricTransactionSignedWithTxHashEnableEpoch = "moa_transaction_signed_with_txhash_enable_epoch"

	// MetricMetaProtectionEnableEpoch represents the epoch when the transactions to the metachain are checked to have enough gas
	MetricMetaProtectionEnableEpoch = "moa_meta_protection_enable_epoch"

	// MetricAheadOfTimeGasUsageEnableEpoch represents the epoch when the cost of smart contract prepare changes from compiler
	// per byte to ahead of time prepare per byte
	MetricAheadOfTimeGasUsageEnableEpoch = "moa_ahead_of_time_gas_usage_enable_epoch"

	// MetricGasPriceModifierEnableEpoch represents the epoch when the gas price modifier in fee computation is enabled
	MetricGasPriceModifierEnableEpoch = "moa_gas_price_modifier_enable_epoch"

	// MetricRepairCallbackEnableEpoch represents the epoch when the callback repair is activated for smart contract results
	MetricRepairCallbackEnableEpoch = "moa_repair_callback_enable_epoch"

	// MetricBlockGasAndFreeRecheckEnableEpoch represents the epoch when gas and fees used in each created or processed block are re-checked
	MetricBlockGasAndFreeRecheckEnableEpoch = "moa_block_gas_and_fee_recheck_enable_epoch"

	// MetricStakingV2EnableEpoch represents the epoch when staking v2 is enabled
	MetricStakingV2EnableEpoch = "moa_staking_v2_enable_epoch"

	// MetricStakeEnableEpoch represents the epoch when staking is enabled
	MetricStakeEnableEpoch = "moa_stake_enable_epoch"

	// MetricDoubleKeyProtectionEnableEpoch represents the epoch when double key protection is enabled
	MetricDoubleKeyProtectionEnableEpoch = "moa_double_key_protection_enable_epoch"

	// MetricDcdtEnableEpoch represents the epoch when DCDT is enabled
	MetricDcdtEnableEpoch = "moa_dcdt_enable_epoch"

	// MetricGovernanceEnableEpoch  represents the epoch when governance is enabled
	MetricGovernanceEnableEpoch = "moa_governance_enable_epoch"

	// MetricDelegationManagerEnableEpoch represents the epoch when the delegation manager is enabled
	MetricDelegationManagerEnableEpoch = "moa_delegation_manager_enable_epoch"

	// MetricDelegationSmartContractEnableEpoch represents the epoch when delegation smart contract is enabled
	MetricDelegationSmartContractEnableEpoch = "moa_delegation_smart_contract_enable_epoch"

	// MetricCorrectLastUnjailedEnableEpoch represents the epoch when the correction on the last unjailed node is applied
	MetricCorrectLastUnjailedEnableEpoch = "moa_correct_last_unjailed_enable_epoch"

	// MetricBalanceWaitingListsEnableEpoch represents the epoch when the balance waiting lists on shards fix is applied
	MetricBalanceWaitingListsEnableEpoch = "moa_balance_waiting_lists_enable_epoch"

	// MetricReturnDataToLastTransferEnableEpoch represents the epoch when the return data to last transfer is applied
	MetricReturnDataToLastTransferEnableEpoch = "moa_return_data_to_last_transfer_enable_epoch"

	// MetricSenderInOutTransferEnableEpoch represents the epoch when the sender in out transfer is applied
	MetricSenderInOutTransferEnableEpoch = "moa_sender_in_out_transfer_enable_epoch"

	// MetricRelayedTransactionsV2EnableEpoch represents the epoch when the relayed transactions v2 is enabled
	MetricRelayedTransactionsV2EnableEpoch = "moa_relayed_transactions_v2_enable_epoch"

	// MetricUnbondTokensV2EnableEpoch represents the epoch when the unbond tokens v2 is applied
	MetricUnbondTokensV2EnableEpoch = "moa_unbond_tokens_v2_enable_epoch"

	// MetricSaveJailedAlwaysEnableEpoch represents the epoch the save jailed fix is applied
	MetricSaveJailedAlwaysEnableEpoch = "moa_save_jailed_always_enable_epoch"

	// MetricValidatorToDelegationEnableEpoch represents the epoch when the validator to delegation feature (staking v3.5) is enabled
	MetricValidatorToDelegationEnableEpoch = "moa_validator_to_delegation_enable_epoch"

	// MetricReDelegateBelowMinCheckEnableEpoch represents the epoch when the re-delegation below minimum value fix is applied
	MetricReDelegateBelowMinCheckEnableEpoch = "moa_redelegate_below_min_check_enable_epoch"

	// MetricIncrementSCRNonceInMultiTransferEnableEpoch represents the epoch when the fix for multi transfer SCR is enabled
	MetricIncrementSCRNonceInMultiTransferEnableEpoch = "moa_increment_scr_nonce_in_multi_transfer_enable_epoch"

	// MetricDCDTMultiTransferEnableEpoch represents the epoch when the DCDT multi transfer feature is enabled
	MetricDCDTMultiTransferEnableEpoch = "moa_dcdt_multi_transfer_enable_epoch"

	// MetricGlobalMintBurnDisableEpoch represents the epoch when the global mint and burn feature is disabled
	MetricGlobalMintBurnDisableEpoch = "moa_global_mint_burn_disable_epoch"

	// MetricDCDTTransferRoleEnableEpoch represents the epoch when the DCDT transfer role feature is enabled
	MetricDCDTTransferRoleEnableEpoch = "moa_dcdt_transfer_role_enable_epoch"

	// MetricMaxNodesChangeEnableEpoch holds configuration for changing the maximum number of nodes and the enabling epoch
	MetricMaxNodesChangeEnableEpoch = "moa_max_nodes_change_enable_epoch"

	// MetricEpochEnable represents the epoch when the max nodes change configuration is applied
	MetricEpochEnable = "moa_epoch_enable"

	// EpochEnableSuffix represents the suffix for EpochEnable item in MaxNodesChangeEnableEpoch list
	EpochEnableSuffix = "_epoch_enable"

	// MetricMaxNumNodes represents the maximum number of nodes than can be enabled in a max nodes change configuration setup
	MetricMaxNumNodes = "moa_max_num_nodes"

	// MaxNumNodesSuffix represents the suffix for MaxNumNodes item in MaxNodesChangeEnableEpoch list
	MaxNumNodesSuffix = "_max_num_nodes"

	// MetricNodesToShufflePerShard represents the nodes to be shuffled per shard
	MetricNodesToShufflePerShard = "moa_nodes_to_shuffle_per_shard"

	// NodesToShufflePerShardSuffix represents the suffix for NodesToShufflePerShard item in MaxNodesChangeEnableEpoch list
	NodesToShufflePerShardSuffix = "_nodes_to_shuffle_per_shard"

	// MetricHysteresis represents the hysteresis threshold
	MetricHysteresis = "moa_hysteresis"

	// MetricAdaptivity represents a boolean to determine if adaptivity will be enabled or not
	MetricAdaptivity = "moa_adaptivity"
)

const (
	// MetricRatingsGeneralStartRating represents the starting rating used by the rater
	MetricRatingsGeneralStartRating = "moa_ratings_general_start_rating"

	// MetricRatingsGeneralMaxRating represents the maximum rating limit
	MetricRatingsGeneralMaxRating = "moa_ratings_general_max_rating"

	// MetricRatingsGeneralMinRating represents the minimum rating limit
	MetricRatingsGeneralMinRating = "moa_ratings_general_min_rating"

	// MetricRatingsGeneralSignedBlocksThreshold represents the signed blocks threshold
	MetricRatingsGeneralSignedBlocksThreshold = "moa_ratings_general_signed_blocks_threshold"

	// MetricRatingsGeneralSelectionChances represents the selection chances thresholds
	MetricRatingsGeneralSelectionChances = "moa_ratings_general_selection_chances"

	// MetricSelectionChancesMaxThreshold represents the max threshold for a selection chances item
	MetricSelectionChancesMaxThreshold = "moa_max_threshold"

	// SelectionChancesMaxThresholdSuffix represents the SelectionChances suffix for MaxThreshold
	SelectionChancesMaxThresholdSuffix = "_max_threshold"

	// MetricSelectionChancesChancePercent represents the chance percentage for a selection chances metric
	MetricSelectionChancesChancePercent = "moa_chance_percent"

	// SelectionChancesChancePercentSuffix represents the SelectionChances suffix for ChancePercent
	SelectionChancesChancePercentSuffix = "_chance_percent"

	// MetricRatingsShardChainHoursToMaxRatingFromStartRating represents the hours to max rating from start rating
	MetricRatingsShardChainHoursToMaxRatingFromStartRating = "moa_ratings_shardchain_hours_to_max_rating_from_start_rating"

	// MetricRatingsShardChainProposerValidatorImportance represents the proposer validator importance index
	MetricRatingsShardChainProposerValidatorImportance = "moa_ratings_shardchain_proposer_validator_importance"

	// MetricRatingsShardChainProposerDecreaseFactor represents the proposer decrease factor
	MetricRatingsShardChainProposerDecreaseFactor = "moa_ratings_shardchain_proposer_decrease_factor"

	// MetricRatingsShardChainValidatorDecreaseFactor represents the validator decrease factor
	MetricRatingsShardChainValidatorDecreaseFactor = "moa_ratings_shardchain_validator_decrease_factor"

	// MetricRatingsShardChainConsecutiveMissedBlocksPenalty represents the consecutive missed block penalty
	MetricRatingsShardChainConsecutiveMissedBlocksPenalty = "moa_ratings_shardchain_consecutive_missed_blocks_penalty"

	// MetricRatingsMetaChainHoursToMaxRatingFromStartRating represents the hours to max rating from start rating
	MetricRatingsMetaChainHoursToMaxRatingFromStartRating = "moa_ratings_metachain_hours_to_max_rating_from_start_rating"

	// MetricRatingsMetaChainProposerValidatorImportance represents the proposer validator importance index
	MetricRatingsMetaChainProposerValidatorImportance = "moa_ratings_metachain_proposer_validator_importance"

	// MetricRatingsMetaChainProposerDecreaseFactor represents the proposer decrease factor
	MetricRatingsMetaChainProposerDecreaseFactor = "moa_ratings_metachain_proposer_decrease_factor"

	// MetricRatingsMetaChainValidatorDecreaseFactor represents the validator decrease factor
	MetricRatingsMetaChainValidatorDecreaseFactor = "moa_ratings_metachain_validator_decrease_factor"

	// MetricRatingsMetaChainConsecutiveMissedBlocksPenalty represents the consecutive missed blocks penalty
	MetricRatingsMetaChainConsecutiveMissedBlocksPenalty = "moa_ratings_metachain_consecutive_missed_blocks_penalty"

	// MetricRatingsPeerHonestyDecayCoefficient represents the peer honesty decay coefficient
	MetricRatingsPeerHonestyDecayCoefficient = "moa_ratings_peerhonesty_decay_coefficient"

	// MetricRatingsPeerHonestyDecayUpdateIntervalInSeconds represents the decat update interval in seconds
	MetricRatingsPeerHonestyDecayUpdateIntervalInSeconds = "moa_ratings_peerhonesty_decay_update_interval_inseconds"

	// MetricRatingsPeerHonestyMaxScore represents the peer honesty max score allowed
	MetricRatingsPeerHonestyMaxScore = "moa_ratings_peerhonesty_max_score"

	// MetricRatingsPeerHonestyMinScore represents the peer honesty min score
	MetricRatingsPeerHonestyMinScore = "moa_ratings_peerhonesty_min_score"

	// MetricRatingsPeerHonestyBadPeerThreshold represents the peer honesty bad peer threshold
	MetricRatingsPeerHonestyBadPeerThreshold = "moa_ratings_peerhonesty_bad_peer_threshold"

	// MetricRatingsPeerHonestyUnitValue represents the peer honesty unit value
	MetricRatingsPeerHonestyUnitValue = "moa_ratings_peerhonesty_unit_value"

	// MetricSetGuardianEnableEpoch represents the epoch when the guardian feature is enabled
	MetricSetGuardianEnableEpoch = "moa_set_guardian_feature_enable_epoch"

	// MetricSetScToScLogEventEnableEpoch represents the epoch when the sc to sc log event feature is enabled
	MetricSetScToScLogEventEnableEpoch = "moa_set_sc_to_sc_log_event_enable_epoch"
)

const (
	// StorerOrder defines the order of storers to be notified of a start of epoch event
	StorerOrder = iota
	// NodesCoordinatorOrder defines the order in which NodesCoordinator is notified of a start of epoch event
	NodesCoordinatorOrder
	// ConsensusOrder defines the order in which Consensus is notified of a start of epoch event
	ConsensusOrder
	// NetworkShardingOrder defines the order in which the network sharding subsystem is notified of a start of epoch event
	NetworkShardingOrder
	// IndexerOrder defines the order in which indexer is notified of a start of epoch event
	IndexerOrder
	// NetStatisticsOrder defines the order in which netStatistic component is notified of a start of epoch event
	NetStatisticsOrder
	// OldDatabaseCleanOrder defines the order in which oldDatabaseCleaner component is notified of a start of epoch event
	OldDatabaseCleanOrder
)

// NodeState specifies what type of state a node could have
type NodeState int

const (
	// NsSynchronized defines ID of a state of synchronized
	NsSynchronized NodeState = iota
	// NsNotSynchronized defines ID of a state of not synchronized
	NsNotSynchronized
	// NsNotCalculated defines ID of a state which is not calculated
	NsNotCalculated
)

// MetricP2PPeerInfo is the metric for the node's p2p info
const MetricP2PPeerInfo = "moa_p2p_peer_info"

// MetricP2PIntraShardValidators is the metric that outputs the intra-shard connected validators
const MetricP2PIntraShardValidators = "moa_p2p_intra_shard_validators"

// MetricP2PCrossShardValidators is the metric that outputs the cross-shard connected validators
const MetricP2PCrossShardValidators = "moa_p2p_cross_shard_validators"

// MetricP2PIntraShardObservers is the metric that outputs the intra-shard connected observers
const MetricP2PIntraShardObservers = "moa_p2p_intra_shard_observers"

// MetricP2PCrossShardObservers is the metric that outputs the cross-shard connected observers
const MetricP2PCrossShardObservers = "moa_p2p_cross_shard_observers"

// MetricP2PUnknownPeers is the metric that outputs the unknown-shard connected peers
const MetricP2PUnknownPeers = "moa_p2p_unknown_shard_peers"

// MetricP2PNumConnectedPeersClassification is the metric for monitoring the number of connected peers split on the connection type
const MetricP2PNumConnectedPeersClassification = "moa_p2p_num_connected_peers_classification"

// MetricAreVMQueriesReady will hold the string representation of the boolean that indicated if the node is ready
// to process VM queries
const MetricAreVMQueriesReady = "moa_are_vm_queries_ready"

// HighestRoundFromBootStorage is the key for the highest round that is saved in storage
const HighestRoundFromBootStorage = "highestRoundFromBootStorage"

// TriggerRegistryKeyPrefix is the key prefix to save epoch start registry to storage
const TriggerRegistryKeyPrefix = "epochStartTrigger_"

// TriggerRegistryInitialKeyPrefix is the key prefix to save initial data to storage
const TriggerRegistryInitialKeyPrefix = "initial_value_epoch_"

// NodesCoordinatorRegistryKeyPrefix is the key prefix to save epoch start registry to storage
const NodesCoordinatorRegistryKeyPrefix = "indexHashed_"

// ShuffledOut signals that a restart is pending because the node was shuffled out
const ShuffledOut = "shuffledOut"

// WrongConfiguration signals that the node has a malformed configuration and cannot continue processing
const WrongConfiguration = "wrongConfiguration"

// ImportComplete signals that a node restart will be done because the import did complete
const ImportComplete = "importComplete"

// DefaultStatsPath is the default path where the node stats are logged
const DefaultStatsPath = "stats"

// DefaultDBPath is the default path for nodes databases
const DefaultDBPath = "db"

// MetachainShardName is the string identifier of the metachain shard
const MetachainShardName = "metachain"

// TemporaryPath is the default temporary path directory
const TemporaryPath = "temp"

// TimeToWaitForP2PBootstrap is the wait time for the P2P to bootstrap
const TimeToWaitForP2PBootstrap = 20 * time.Second

// MaxSoftwareVersionLengthInBytes represents the maximum length for the software version to be saved in block header
const MaxSoftwareVersionLengthInBytes = 10

// ExtraDelayForBroadcastBlockInfo represents the number of seconds to wait since a block has been broadcast and the
// moment when its components, like mini blocks and transactions, would be broadcast too
const ExtraDelayForBroadcastBlockInfo = 1 * time.Second

// ExtraDelayBetweenBroadcastMbsAndTxs represents the number of seconds to wait since miniblocks have been broadcast
// and the moment when theirs transactions would be broadcast too
const ExtraDelayBetweenBroadcastMbsAndTxs = 1 * time.Second

// ExtraDelayForRequestBlockInfo represents the number of seconds to wait since a block has been received and the
// moment when its components, like mini blocks and transactions, would be requested too if they are still missing
const ExtraDelayForRequestBlockInfo = ExtraDelayForBroadcastBlockInfo + ExtraDelayBetweenBroadcastMbsAndTxs + time.Second

// CommitMaxTime represents max time accepted for a commit action, after which a warn message is displayed
const CommitMaxTime = 3 * time.Second

// PutInStorerMaxTime represents max time accepted for a put action, after which a warn message is displayed
const PutInStorerMaxTime = time.Second

// DefaultUnstakedEpoch represents the default epoch that is set for a validator that has not unstaked yet
const DefaultUnstakedEpoch = math.MaxUint32

// InvalidMessageBlacklistDuration represents the time to keep a peer in the black list if it sends a message that
// does not follow the protocol: example not useing the same marshaler as the other peers
const InvalidMessageBlacklistDuration = time.Second * 3600

// PublicKeyBlacklistDuration represents the time to keep a public key in the black list if it will degrade its
// rating to a minimum threshold due to improper messages
const PublicKeyBlacklistDuration = time.Second * 7200

// InvalidSigningBlacklistDuration defines the time to keep a peer id in blacklist if it signs a message with invalid signature
const InvalidSigningBlacklistDuration = time.Second * 7200

// MaxWaitingTimeToReceiveRequestedItem represents the maximum waiting time in seconds needed to receive the requested items
const MaxWaitingTimeToReceiveRequestedItem = 5 * time.Second

// DefaultLogProfileIdentifier represents the default log profile used when the logviewer/termui applications do not
// need to change the current logging profile
const DefaultLogProfileIdentifier = "[default log profile]"

// NotSetDestinationShardID represents the shardIdString when the destinationShardId is not set in the prefs
const NotSetDestinationShardID = "disabled"

// AdditionalScrForEachScCallOrSpecialTx specifies the additional number of smart contract results which should be
// considered by a node, when it includes sc calls or special txs in a miniblock.
// Ex.: normal txs -> aprox. 27000, sc calls or special txs -> aprox. 6250 = 27000 / (AdditionalScrForEachScCallOrSpecialTx + 1),
// considering that constant below is set to 3
const AdditionalScrForEachScCallOrSpecialTx = 3

// MaxRoundsWithoutCommittedStartInEpochBlock defines the maximum rounds to wait for start in epoch block to be committed,
// before a special action to be applied
const MaxRoundsWithoutCommittedStartInEpochBlock = 50

// DefaultResolversIdentifier represents the identifier that is used in conjunction with regular resolvers
// (that makes the node run properly)
const DefaultResolversIdentifier = "default resolver"

// DefaultInterceptorsIdentifier represents the identifier that is used in conjunction with regular interceptors
// (that makes the node run properly)
const DefaultInterceptorsIdentifier = "default interceptor"

// HardforkInterceptorsIdentifier represents the identifier that is used in the hardfork process
const HardforkInterceptorsIdentifier = "hardfork interceptor"

// HardforkResolversIdentifier represents the resolver that is used in the hardfork process
const HardforkResolversIdentifier = "hardfork resolver"

// EpochStartInterceptorsIdentifier represents the identifier that is used in the start-in-epoch process
const EpochStartInterceptorsIdentifier = "epoch start interceptor"

// TimeoutGettingTrieNodes defines the timeout in trie sync operation if no node is received
const TimeoutGettingTrieNodes = 2 * time.Minute // to consider syncing a very large trie node of 64MB at ~1MB/s

// TimeoutGettingTrieNodesInHardfork represents the maximum time allowed between 2 nodes fetches (and commits)
// during the hardfork process
const TimeoutGettingTrieNodesInHardfork = time.Minute * 10

// RetrialIntervalForOutportDriver is the interval in which the outport driver should try to call the driver again
const RetrialIntervalForOutportDriver = time.Second * 10

// NodeProcessingMode represents the processing mode in which the node was started
type NodeProcessingMode int

const (
	// Normal means that the node has started in the normal processing mode
	Normal NodeProcessingMode = iota

	// ImportDb means that the node has started in the import-db mode
	ImportDb
)

const (
	// ActiveDBKey is the key at which ActiveDBVal will be saved
	ActiveDBKey = "activeDB"

	// ActiveDBVal is the value that will be saved at ActiveDBKey
	ActiveDBVal = "yes"

	// TrieSyncedKey is the key at which TrieSyncedVal will be saved
	TrieSyncedKey = "synced"

	// TrieSyncedVal is the value that will be saved at TrieSyncedKey
	TrieSyncedVal = "yes"

	// TrieLeavesChannelDefaultCapacity represents the default value to be used as capacity for getting all trie leaves on
	// a channel
	TrieLeavesChannelDefaultCapacity = 100

	// TrieLeavesChannelSyncCapacity represents the value to be used as capacity for getting main trie
	// leaf nodes for trie sync
	TrieLeavesChannelSyncCapacity = 1000
)

// ApiOutputFormat represents the format type returned by api
type ApiOutputFormat uint8

const (
	// ApiOutputFormatJSON outport format returns struct directly, will be serialized into JSON by gin
	ApiOutputFormatJSON ApiOutputFormat = 0

	// ApiOutputFormatProto outport format returns the bytes of the proto object
	ApiOutputFormatProto ApiOutputFormat = 1
)

// BlockProcessingCutoffMode represents the type to be used to identify the mode of the block processing cutoff
type BlockProcessingCutoffMode string

const (
	// BlockProcessingCutoffModePause represents the mode where the node will pause the processing at the given coordinates
	BlockProcessingCutoffModePause = "pause"
	// BlockProcessingCutoffModeProcessError represents the mode where the node will reprocess with error the block at the given coordinates
	BlockProcessingCutoffModeProcessError = "process-error"
)

// BlockProcessingCutoffTrigger represents the trigger of the cutoff potentially used in block processing
type BlockProcessingCutoffTrigger string

const (
	// BlockProcessingCutoffByNonce represents the cutoff by nonce
	BlockProcessingCutoffByNonce BlockProcessingCutoffTrigger = "nonce"
	// BlockProcessingCutoffByRound represents the cutoff by round
	BlockProcessingCutoffByRound BlockProcessingCutoffTrigger = "round"
	// BlockProcessingCutoffByEpoch represents the cutoff by epoch
	BlockProcessingCutoffByEpoch BlockProcessingCutoffTrigger = "epoch"
)

// MaxIndexOfTxInMiniBlock defines the maximum index of a tx inside one mini block
const MaxIndexOfTxInMiniBlock = int32(29999)

// MetricAccountsSnapshotInProgress is the metric that outputs the status of the accounts' snapshot, if it's in progress or not
const MetricAccountsSnapshotInProgress = "moa_accounts_snapshot_in_progress"

// MetricLastAccountsSnapshotDurationSec is the metric that outputs the duration in seconds of the last accounts db snapshot. If snapshot is in progress it will be set to 0
const MetricLastAccountsSnapshotDurationSec = "moa_accounts_snapshot_last_duration_in_seconds"

// MetricPeersSnapshotInProgress is the metric that outputs the status of the peers' snapshot, if it's in progress or not
const MetricPeersSnapshotInProgress = "moa_peers_snapshot_in_progress"

// MetricLastPeersSnapshotDurationSec is the metric that outputs the duration in seconds of the last peers db snapshot. If snapshot is in progress it will be set to 0
const MetricLastPeersSnapshotDurationSec = "moa_peers_snapshot_last_duration_in_seconds"

// GenesisStorageSuffix defines the storage suffix used for genesis altered data
const GenesisStorageSuffix = "_genesis"

// MetricAccountsSnapshotNumNodes is the metric that outputs the number of trie nodes written for accounts after snapshot
const MetricAccountsSnapshotNumNodes = "moa_accounts_snapshot_num_nodes"

// MetricTrieSyncNumReceivedBytes is the metric that outputs the number of bytes received for accounts during trie sync
const MetricTrieSyncNumReceivedBytes = "moa_trie_sync_num_bytes_received"

// MetricTrieSyncNumProcessedNodes is the metric that outputs the number of trie nodes processed for accounts during trie sync
const MetricTrieSyncNumProcessedNodes = "moa_trie_sync_num_nodes_processed"

// FullArchiveMetricSuffix is the suffix added to metrics specific for full archive network
const FullArchiveMetricSuffix = "_full_archive"

// Enable epoch flags definitions
const (
	SCDeployFlag                                       core.EnableEpochFlag = "SCDeployFlag"
	BuiltInFunctionsFlag                               core.EnableEpochFlag = "BuiltInFunctionsFlag"
	RelayedTransactionsFlag                            core.EnableEpochFlag = "RelayedTransactionsFlag"
	PenalizedTooMuchGasFlag                            core.EnableEpochFlag = "PenalizedTooMuchGasFlag"
	SwitchJailWaitingFlag                              core.EnableEpochFlag = "SwitchJailWaitingFlag"
	BelowSignedThresholdFlag                           core.EnableEpochFlag = "BelowSignedThresholdFlag"
	SwitchHysteresisForMinNodesFlagInSpecificEpochOnly core.EnableEpochFlag = "SwitchHysteresisForMinNodesFlagInSpecificEpochOnly"
	TransactionSignedWithTxHashFlag                    core.EnableEpochFlag = "TransactionSignedWithTxHashFlag"
	MetaProtectionFlag                                 core.EnableEpochFlag = "MetaProtectionFlag"
	AheadOfTimeGasUsageFlag                            core.EnableEpochFlag = "AheadOfTimeGasUsageFlag"
	GasPriceModifierFlag                               core.EnableEpochFlag = "GasPriceModifierFlag"
	RepairCallbackFlag                                 core.EnableEpochFlag = "RepairCallbackFlag"
	ReturnDataToLastTransferFlagAfterEpoch             core.EnableEpochFlag = "ReturnDataToLastTransferFlagAfterEpoch"
	SenderInOutTransferFlag                            core.EnableEpochFlag = "SenderInOutTransferFlag"
	StakeFlag                                          core.EnableEpochFlag = "StakeFlag"
	StakingV2Flag                                      core.EnableEpochFlag = "StakingV2Flag"
	StakingV2OwnerFlagInSpecificEpochOnly              core.EnableEpochFlag = "StakingV2OwnerFlagInSpecificEpochOnly"
	StakingV2FlagAfterEpoch                            core.EnableEpochFlag = "StakingV2FlagAfterEpoch"
	DoubleKeyProtectionFlag                            core.EnableEpochFlag = "DoubleKeyProtectionFlag"
	DCDTFlag                                           core.EnableEpochFlag = "DCDTFlag"
	DCDTFlagInSpecificEpochOnly                        core.EnableEpochFlag = "DCDTFlagInSpecificEpochOnly"
	GovernanceFlag                                     core.EnableEpochFlag = "GovernanceFlag"
	GovernanceFlagInSpecificEpochOnly                  core.EnableEpochFlag = "GovernanceFlagInSpecificEpochOnly"
	DelegationManagerFlag                              core.EnableEpochFlag = "DelegationManagerFlag"
	DelegationSmartContractFlag                        core.EnableEpochFlag = "DelegationSmartContractFlag"
	DelegationSmartContractFlagInSpecificEpochOnly     core.EnableEpochFlag = "DelegationSmartContractFlagInSpecificEpochOnly"
	CorrectLastUnJailedFlag                            core.EnableEpochFlag = "CorrectLastUnJailedFlag"
	CorrectLastUnJailedFlagInSpecificEpochOnly         core.EnableEpochFlag = "CorrectLastUnJailedFlagInSpecificEpochOnly"
	RelayedTransactionsV2Flag                          core.EnableEpochFlag = "RelayedTransactionsV2Flag"
	UnBondTokensV2Flag                                 core.EnableEpochFlag = "UnBondTokensV2Flag"
	SaveJailedAlwaysFlag                               core.EnableEpochFlag = "SaveJailedAlwaysFlag"
	ReDelegateBelowMinCheckFlag                        core.EnableEpochFlag = "ReDelegateBelowMinCheckFlag"
	ValidatorToDelegationFlag                          core.EnableEpochFlag = "ValidatorToDelegationFlag"
	IncrementSCRNonceInMultiTransferFlag               core.EnableEpochFlag = "IncrementSCRNonceInMultiTransferFlag"
	DCDTMultiTransferFlag                              core.EnableEpochFlag = "DCDTMultiTransferFlag"
	GlobalMintBurnFlag                                 core.EnableEpochFlag = "GlobalMintBurnFlag"
	DCDTTransferRoleFlag                               core.EnableEpochFlag = "DCDTTransferRoleFlag"
	ComputeRewardCheckpointFlag                        core.EnableEpochFlag = "ComputeRewardCheckpointFlag"
	SCRSizeInvariantCheckFlag                          core.EnableEpochFlag = "SCRSizeInvariantCheckFlag"
	BackwardCompSaveKeyValueFlag                       core.EnableEpochFlag = "BackwardCompSaveKeyValueFlag"
	DCDTNFTCreateOnMultiShardFlag                      core.EnableEpochFlag = "DCDTNFTCreateOnMultiShardFlag"
	MetaDCDTSetFlag                                    core.EnableEpochFlag = "MetaDCDTSetFlag"
	AddTokensToDelegationFlag                          core.EnableEpochFlag = "AddTokensToDelegationFlag"
	MultiDCDTTransferFixOnCallBackFlag                 core.EnableEpochFlag = "MultiDCDTTransferFixOnCallBackFlag"
	OptimizeGasUsedInCrossMiniBlocksFlag               core.EnableEpochFlag = "OptimizeGasUsedInCrossMiniBlocksFlag"
	CorrectFirstQueuedFlag                             core.EnableEpochFlag = "CorrectFirstQueuedFlag"
	DeleteDelegatorAfterClaimRewardsFlag               core.EnableEpochFlag = "DeleteDelegatorAfterClaimRewardsFlag"
	RemoveNonUpdatedStorageFlag                        core.EnableEpochFlag = "RemoveNonUpdatedStorageFlag"
	OptimizeNFTStoreFlag                               core.EnableEpochFlag = "OptimizeNFTStoreFlag"
	CreateNFTThroughExecByCallerFlag                   core.EnableEpochFlag = "CreateNFTThroughExecByCallerFlag"
	StopDecreasingValidatorRatingWhenStuckFlag         core.EnableEpochFlag = "StopDecreasingValidatorRatingWhenStuckFlag"
	FrontRunningProtectionFlag                         core.EnableEpochFlag = "FrontRunningProtectionFlag"
	PayableBySCFlag                                    core.EnableEpochFlag = "PayableBySCFlag"
	CleanUpInformativeSCRsFlag                         core.EnableEpochFlag = "CleanUpInformativeSCRsFlag"
	StorageAPICostOptimizationFlag                     core.EnableEpochFlag = "StorageAPICostOptimizationFlag"
	DCDTRegisterAndSetAllRolesFlag                     core.EnableEpochFlag = "DCDTRegisterAndSetAllRolesFlag"
	ScheduledMiniBlocksFlag                            core.EnableEpochFlag = "ScheduledMiniBlocksFlag"
	CorrectJailedNotUnStakedEmptyQueueFlag             core.EnableEpochFlag = "CorrectJailedNotUnStakedEmptyQueueFlag"
	DoNotReturnOldBlockInBlockchainHookFlag            core.EnableEpochFlag = "DoNotReturnOldBlockInBlockchainHookFlag"
	AddFailedRelayedTxToInvalidMBsFlag                 core.EnableEpochFlag = "AddFailedRelayedTxToInvalidMBsFlag"
	SCRSizeInvariantOnBuiltInResultFlag                core.EnableEpochFlag = "SCRSizeInvariantOnBuiltInResultFlag"
	CheckCorrectTokenIDForTransferRoleFlag             core.EnableEpochFlag = "CheckCorrectTokenIDForTransferRoleFlag"
	FailExecutionOnEveryAPIErrorFlag                   core.EnableEpochFlag = "FailExecutionOnEveryAPIErrorFlag"
	MiniBlockPartialExecutionFlag                      core.EnableEpochFlag = "MiniBlockPartialExecutionFlag"
	ManagedCryptoAPIsFlag                              core.EnableEpochFlag = "ManagedCryptoAPIsFlag"
	DCDTMetadataContinuousCleanupFlag                  core.EnableEpochFlag = "DCDTMetadataContinuousCleanupFlag"
	DisableExecByCallerFlag                            core.EnableEpochFlag = "DisableExecByCallerFlag"
	RefactorContextFlag                                core.EnableEpochFlag = "RefactorContextFlag"
	CheckFunctionArgumentFlag                          core.EnableEpochFlag = "CheckFunctionArgumentFlag"
	CheckExecuteOnReadOnlyFlag                         core.EnableEpochFlag = "CheckExecuteOnReadOnlyFlag"
	SetSenderInEeiOutputTransferFlag                   core.EnableEpochFlag = "SetSenderInEeiOutputTransferFlag"
	FixAsyncCallbackCheckFlag                          core.EnableEpochFlag = "FixAsyncCallbackCheckFlag"
	SaveToSystemAccountFlag                            core.EnableEpochFlag = "SaveToSystemAccountFlag"
	CheckFrozenCollectionFlag                          core.EnableEpochFlag = "CheckFrozenCollectionFlag"
	SendAlwaysFlag                                     core.EnableEpochFlag = "SendAlwaysFlag"
	ValueLengthCheckFlag                               core.EnableEpochFlag = "ValueLengthCheckFlag"
	CheckTransferFlag                                  core.EnableEpochFlag = "CheckTransferFlag"
	DCDTNFTImprovementV1Flag                           core.EnableEpochFlag = "DCDTNFTImprovementV1Flag"
	ChangeDelegationOwnerFlag                          core.EnableEpochFlag = "ChangeDelegationOwnerFlag"
	RefactorPeersMiniBlocksFlag                        core.EnableEpochFlag = "RefactorPeersMiniBlocksFlag"
	SCProcessorV2Flag                                  core.EnableEpochFlag = "SCProcessorV2Flag"
	FixAsyncCallBackArgsListFlag                       core.EnableEpochFlag = "FixAsyncCallBackArgsListFlag"
	FixOldTokenLiquidityFlag                           core.EnableEpochFlag = "FixOldTokenLiquidityFlag"
	RuntimeMemStoreLimitFlag                           core.EnableEpochFlag = "RuntimeMemStoreLimitFlag"
	RuntimeCodeSizeFixFlag                             core.EnableEpochFlag = "RuntimeCodeSizeFixFlag"
	MaxBlockchainHookCountersFlag                      core.EnableEpochFlag = "MaxBlockchainHookCountersFlag"
	WipeSingleNFTLiquidityDecreaseFlag                 core.EnableEpochFlag = "WipeSingleNFTLiquidityDecreaseFlag"
	AlwaysSaveTokenMetaDataFlag                        core.EnableEpochFlag = "AlwaysSaveTokenMetaDataFlag"
	SetGuardianFlag                                    core.EnableEpochFlag = "SetGuardianFlag"
	RelayedNonceFixFlag                                core.EnableEpochFlag = "RelayedNonceFixFlag"
	ConsistentTokensValuesLengthCheckFlag              core.EnableEpochFlag = "ConsistentTokensValuesLengthCheckFlag"
	KeepExecOrderOnCreatedSCRsFlag                     core.EnableEpochFlag = "KeepExecOrderOnCreatedSCRsFlag"
	MultiClaimOnDelegationFlag                         core.EnableEpochFlag = "MultiClaimOnDelegationFlag"
	ChangeUsernameFlag                                 core.EnableEpochFlag = "ChangeUsernameFlag"
	AutoBalanceDataTriesFlag                           core.EnableEpochFlag = "AutoBalanceDataTriesFlag"
	MigrateDataTrieFlag                                core.EnableEpochFlag = "MigrateDataTrieFlag"
	FixDelegationChangeOwnerOnAccountFlag              core.EnableEpochFlag = "FixDelegationChangeOwnerOnAccountFlag"
	FixOOGReturnCodeFlag                               core.EnableEpochFlag = "FixOOGReturnCodeFlag"
	DeterministicSortOnValidatorsInfoFixFlag           core.EnableEpochFlag = "DeterministicSortOnValidatorsInfoFixFlag"
	DynamicGasCostForDataTrieStorageLoadFlag           core.EnableEpochFlag = "DynamicGasCostForDataTrieStorageLoadFlag"
	ScToScLogEventFlag                                 core.EnableEpochFlag = "ScToScLogEventFlag"
	BlockGasAndFeesReCheckFlag                         core.EnableEpochFlag = "BlockGasAndFeesReCheckFlag"
	BalanceWaitingListsFlag                            core.EnableEpochFlag = "BalanceWaitingListsFlag"
	NFTStopCreateFlag                                  core.EnableEpochFlag = "NFTStopCreateFlag"
	FixGasRemainingForSaveKeyValueFlag                 core.EnableEpochFlag = "FixGasRemainingForSaveKeyValueFlag"
	IsChangeOwnerAddressCrossShardThroughSCFlag        core.EnableEpochFlag = "IsChangeOwnerAddressCrossShardThroughSCFlag"
	CurrentRandomnessOnSortingFlag                     core.EnableEpochFlag = "CurrentRandomnessOnSortingFlag"
	StakeLimitsFlag                                    core.EnableEpochFlag = "StakeLimitsFlag"
	StakingV4Step1Flag                                 core.EnableEpochFlag = "StakingV4Step1Flag"
	StakingV4Step2Flag                                 core.EnableEpochFlag = "StakingV4Step2Flag"
	StakingV4Step3Flag                                 core.EnableEpochFlag = "StakingV4Step3Flag"
	CleanupAuctionOnLowWaitingListFlag                 core.EnableEpochFlag = "CleanupAuctionOnLowWaitingListFlag"
	StakingV4StartedFlag                               core.EnableEpochFlag = "StakingV4StartedFlag"
	AlwaysMergeContextsInEEIFlag                       core.EnableEpochFlag = "AlwaysMergeContextsInEEIFlag"
	UseGasBoundedShouldFailExecutionFlag               core.EnableEpochFlag = "UseGasBoundedShouldFailExecutionFlag"
	// all new flags must be added to createAllFlagsMap method, as part of enableEpochsHandler allFlagsDefined
)
