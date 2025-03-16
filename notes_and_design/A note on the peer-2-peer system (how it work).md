// Resolving the cyclic counter-y problems:
// - Note that the problem actually occurs very rarely, as we consider every node
//   to be a single source of truth for its own cab and hall calls
// - The *only* case where a node is not assumed to be correct is if it has crashed,
//   then we need to assume that it has cab calls which it doesn't know, but should
//   be informed about from the other nodes on the network
// - This means that we need to detect when a node has crashed so we can know that the
//   node should have its cab calls overwritten, rather than simply broadcasting "I
//   have zero cab calls"
// - To do this, introduce uptime. Then every backed up request is tagged with the
//   uptime of the node when it was implemented
// - If a node disconnects, its timer will have increased, and so when another node
//   attempts to return the cab calls it will notice that the counter of the node
//   is higher than that of the request, and therefore discard it
// - The node will then broadcast these, and since the uptime value will be larger,
//   the node will overwrite (Take the UNION!!!) with its view of the other node(s)
// - If the node crashes, however, it will return with a lower lifetime than it had
//   before. Then the node will be informed about its lost cab requests, and take these
// - Then these calls are accepted, the node will start to to broadcast it, and then
//   the backups will update the timestamp to be the current timestamp of the node,
//   so when the node is done it will be considered new information and thus it will
//   overwrite the backed up requests.

// A problematic case?
// - Node A dies, and node B and C are left on the network
// - Node B dies and comes back -> (Node B has no backup of A?)
// - Node C dies and comes back -> (Node C has no backup of A?)

// To resolve this, we have the functionality:
// - If someone has a more recent view backup of the node than we do, we update our
//   backup
// - This way everyone must have the most recent view of the node (under reasonable
//   assumptions), so the case is solved.

// Regarding the arbiter/who should redistribute issue:
// - When a peer is lost, the node looks at the updated peer list
// - If there is a peer with an ID with a lower uptime, do nothing
// - Else, redistribute the hall calls that have been backed up