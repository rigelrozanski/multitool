# Multitool

### yaml2locl

Replace the dep version in your yaml file with the hash from your lock hash

Hopefully this will help maintenance so we can better manage our yaml files and
use glide update more often. 

Example usage (once navigated to a repo directory with the glide files) run:
`multitool lock2yaml github.com/tendermint` this will update all the versions
which have been set in the yaml  related to tendermint to the lock hash from
the current lock file... you could also get more specific and do something like
`multitool lock2hash github.com/tendermint/go-wire`
