package repo

// Config mirrors governance/config.yaml. Only the fields the checks use are typed.
type Config struct {
	Review struct {
		ReviewsRequired          int     `yaml:"reviews_required"`
		MinReviewerModelFamilies int     `yaml:"min_reviewer_model_families"`
		ReviewsOwedPerSubmission int     `yaml:"reviews_owed_per_submission"`
		ApproveThresholdWeighted float64 `yaml:"approve_threshold_weighted"`
	} `yaml:"review"`
	Crosswalks struct {
		FingerprintMatchThreshold float64 `yaml:"fingerprint_match_threshold"`
		CrosswalkMinSightingsOrgs int     `yaml:"crosswalk_min_sightings_orgs"`
	} `yaml:"crosswalks"`
	Drafts struct {
		DraftExpiryMonths int     `yaml:"draft_expiry_months"`
		OverlapThreshold  float64 `yaml:"overlap_threshold"`
	} `yaml:"drafts"`
	Evidence struct {
		MinTierForSpecChange             string `yaml:"min_tier_for_spec_change"`
		MinTierForCrosswalkRequiredFields string `yaml:"min_tier_for_crosswalk_required_fields"`
	} `yaml:"evidence"`
	Privacy struct {
		SyntheticRules   string   `yaml:"synthetic_rules"`
		ScrubExemptPaths []string `yaml:"scrub_exempt_paths"`
	} `yaml:"privacy"`
	Snapshot struct {
		MaxBytes       int64    `yaml:"max_bytes"`
		StaleAfterDays int      `yaml:"stale_after_days"`
		Include        []string `yaml:"include"`
		Exclude        []string `yaml:"exclude"`
	} `yaml:"snapshot"`
}

// Industry is one entry of governance/industries.yaml.
type Industry struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	RDC  string `yaml:"rdc"`
}

// TestArea is a designated offshore box for synthetic coordinates.
type TestArea struct {
	ID          string  `yaml:"id"`
	Description string  `yaml:"description"`
	MinLon      float64 `yaml:"min_lon"`
	MaxLon      float64 `yaml:"max_lon"`
	MinLat      float64 `yaml:"min_lat"`
	MaxLat      float64 `yaml:"max_lat"`
}

// Synthetic mirrors governance/synthetic.yaml.
type Synthetic struct {
	TestAreas []TestArea `yaml:"test_areas"`
}

// Config loads governance/config.yaml.
func (r *Repo) Config() (*Config, error) {
	var c Config
	return &c, r.ReadConfigInto("governance/config.yaml", &c)
}

// Industries loads governance/industries.yaml.
func (r *Repo) Industries() ([]Industry, error) {
	var v struct {
		Industries []Industry `yaml:"industries"`
	}
	err := r.ReadConfigInto("governance/industries.yaml", &v)
	return v.Industries, err
}

// IndustryIDs returns the set of governed industry ids.
func (r *Repo) IndustryIDs() (map[string]bool, error) {
	inds, err := r.Industries()
	if err != nil {
		return nil, err
	}
	m := map[string]bool{}
	for _, i := range inds {
		m[i.ID] = true
	}
	return m, nil
}

// Synthetic loads governance/synthetic.yaml.
func (r *Repo) Synthetic() (*Synthetic, error) {
	var s Synthetic
	return &s, r.ReadConfigInto("governance/synthetic.yaml", &s)
}
