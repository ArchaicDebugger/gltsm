package dbmodels

// User represents the structure of the users table
type User struct {
	ID       string `gorm:"primaryKey"` // Use Mbid as primary key
	Username string `gorm:"unique;not null"`
}

// Artist represents the structure of the artists table
type Artist struct {
	Mbid string `gorm:"primaryKey"` // Use Mbid as primary key
	Name string `gorm:"not null"`
}

// Album represents the structure of the albums table
type Album struct {
	Mbid string `gorm:"primaryKey"` // Use Mbid as primary key
	Name string `gorm:"not null"`
}

// Track represents the structure of the tracks table
type Track struct {
	Mbid       string  `gorm:"primaryKey"` // Use Mbid as primary key
	Name       string  `gorm:"not null"`
	ArtistID   string  `gorm:"not null"` // Foreign key for Artist (Mbid)
	AlbumID    string  `gorm:"not null"` // Foreign key for Album (Mbid)
	Streamable string  `gorm:"not null"`
	URL        string  `gorm:"not null"`
	Date       string  `gorm:"not null"`           // Store as a string or time.Time
	Images     []Image `gorm:"foreignKey:TrackID"` // One-to-many relationship with Images
}

// Image represents the structure of track images
type Image struct {
	ID      uint   `gorm:"primaryKey"`
	TrackID string `gorm:"not null"` // Foreign key for Track (Mbid)
	Size    string `gorm:"not null"`
	URL     string `gorm:"not null"`
}

// ListeningHistory represents the structure of the listening history table
type ListeningHistory struct {
	ID      uint   `gorm:"primaryKey"` // Auto-increment ID
	UserID  string `gorm:"not null"`   // Foreign key for User (ID)
	TrackID string `gorm:"not null"`   // Foreign key for Track (Mbid)
	Date    string `gorm:"not null"`   // Store as a string or time.Time
}
