package grafana

type Dashboard struct {
	Annotations          Annotations `json:"annotations"`
	Editable             bool        `json:"editable"`
	FiscalYearStartMonth int         `json:"fiscalYearStartMonth"`
	GraphTooltip         int         `json:"graphTooltip"`
	ID                   int         `json:"id"`
	Links                []string    `json:"links"`
	LiveNow              bool        `json:"liveNow"`
	Panels               []Panel     `json:"panels"`
	SchemaVersion        int         `json:"schemaVersion"`
	Style                string      `json:"style"`
	Tags                 []string    `json:"tags"`
	Templating           Templating  `json:"templating"`
	Time                 Time        `json:"time"`
	Timepicker           Timepicker  `json:"timepicker"`
	Timezone             string      `json:"timezone"`
	Title                string      `json:"title"`
	UID                  string      `json:"uid"`
	Version              int         `json:"version"`
	WeekStart            string      `json:"weekStart"`
}

type Annotations struct {
	List []AnnotationItem `json:"list"`
}

type AnnotationItem struct {
	BuiltIn    int        `json:"builtIn"`
	Datasource Datasource `json:"datasource"`
	Enable     bool       `json:"enable"`
	Hide       bool       `json:"hide"`
	IconColor  string     `json:"iconColor"`
	Name       string     `json:"name"`
	Target     Target     `json:"target"`
	Type       string     `json:"type"`
}

type Datasource struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
}

type Target struct {
	Limit    int      `json:"limit"`
	MatchAny bool     `json:"matchAny"`
	Tags     []string `json:"tags"`
	Type     string   `json:"type"`
}

type TargetAlt struct {
	Datasource Datasource `json:"datasource"`
	EditorMode string     `json:"editorMode"`
	Exemplar   bool       `json:"exemplar"`
	Expr       string     `json:"expr"`
	Range      bool       `json:"range"`
	RefID      string     `json:"refID"`
}

type Panel struct {
	Collapsed     bool        `json:"collapsed,omitempty"`
	GridPos       GridPos     `json:"gridPos"`
	ID            int         `json:"id"`
	Panels        []Panel     `json:"panels,omitempty"`
	Title         string      `json:"title"`
	Type          string      `json:"type"`
	Datasource    Datasource  `json:"datasource,omitempty"`
	Description   string      `json:"description,omitempty"`
	FieldConfig   FieldConfig `json:"fieldConfig,omitempty"`
	Options       Options     `json:"options,omitempty"`
	PluginVersion string      `json:"pluginVersion,omitempty"`
	Targets       []TargetAlt `json:"targets,omitempty"`
}

type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

type FieldConfig struct {
	Defaults  Defaults `json:"defaults"`
	Overrides []string `json:"overrides"`
}

type Defaults struct {
	Color       Color      `json:"color"`
	DisplayName string     `json:"displayName"`
	Mappings    []string   `json:"mappings"`
	Thresholds  Thresholds `json:"thresholds"`
	Unit        string     `json:"unit"`
}

type Color struct {
	Mode string `json:"mode"`
}

type Thresholds struct {
	Mode  string `json:"mode"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Color string `json:"color"`
	Value int    `json:"value"`
}

type Options struct {
	ColorMode     string        `json:"colorMode"`
	GraphMode     string        `json:"graphMode"`
	JustifyMode   string        `json:"justifyMode"`
	Orientation   string        `json:"orientation"`
	ReduceOptions ReduceOptions `json:"reduceOptions"`
	Text          Text          `json:"text"`
	TextMode      string        `json:"textMode"`
}

type ReduceOptions struct {
	Calcs  []string `json:"calcs"`
	Fields string   `json:"fields"`
	Values bool     `json:"values"`
}

type Text struct {
}

type Templating struct {
	List []string `json:"list"`
}

type Time struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Timepicker struct {
}
