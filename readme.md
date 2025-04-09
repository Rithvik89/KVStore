* Version 1.0 : 
    * Basic Post operations onto read replicas. This may be the go to way of implementing, but this lacks in consistency.
    * Maintaining a WAL file with versioning (for durability)
    * Adding replicas on fly to deal with load.