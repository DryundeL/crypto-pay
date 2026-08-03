package domain

type AccountID string

func (id AccountID) String() string { return string(id) }
func (id AccountID) IsZero() bool   { return id == "" }

type JournalID string

func (id JournalID) String() string { return string(id) }
func (id JournalID) IsZero() bool   { return id == "" }

type EntryID string

func (id EntryID) String() string { return string(id) }
func (id EntryID) IsZero() bool   { return id == "" }
