package svg

const (
	NameOnly       = "nameOnly"     // name only
	DateOnly       = "dateOnly"     // date only
	NameWithDate   = "nameWithDate" // name and date
	barHeight      = 40             // includes 10px spacer
	chartWidth     = 1440           // Number of minutes in a day
//	slotsPerDay    = 96             // Number of 15-minute timeslots in a day
	slotPixelWidth = 15             // 1440 minutes/96 slots
	headerHeight   = 50             // Some room for the title & time markers
	timeLabelY     = 25             // How far down should the time label be
	onlineColor    = "OliveDrab"    // HTML Color Names
	offlineColor   = "LightGray"    // HTML Color Names
	outageColor    = "Pink"         // HTML Color Names
	offline        = 0              // Device offline
	online         = 1              // Device online
	outage         = -1             // System outage
	svgStyle       = `<style>
.titleText {
	font-family: 'Arial', 'Helvetica', sans-serif;
	font-size: 20px;
	font-weight: normal;
	fill: black;
	text-anchor: middle;
}
.whiteText {
	font-family: 'Arial', 'Helvetica', sans-serif;
	font-size: 20px;
	font-weight: normal;
	fill: white;
	text-anchor: left;
}
.blackText {
	font-family: 'Arial', 'Helvetica', sans-serif;
	font-size: 20px;
	font-weight: normal;
	fill: black;
	text-anchor: left;
}
</style>`
)
