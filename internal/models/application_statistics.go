package models

// DailySubmissionCount represents the number of applications submitted on a date.
type DailySubmissionCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// ApplicationStatistics represents aggregate application metrics for the admin dashboard.
type ApplicationStatistics struct {
	TotalApplications      int                    `json:"totalApplications"`
	ReviewedApplications   int                    `json:"reviewedApplications"`
	UnreviewedApplications int                    `json:"unreviewedApplications"`
	AcceptedApplications   int                    `json:"acceptedApplications"`
	PendingApplications    int                    `json:"pendingApplications"`
	RejectedApplications   int                    `json:"rejectedApplications"`
	JoinedDiscordCount     int                    `json:"joinedDiscordCount"`
	AverageAge             float64                `json:"averageAge"`
	RecentSubmissions      []DailySubmissionCount `json:"recentSubmissions"`
}
