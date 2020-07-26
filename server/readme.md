export PATH=$PATH:$(dirname $(go list -f '{{.Target}}' .))
