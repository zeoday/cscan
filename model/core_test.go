package model

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==================== Generators for Property Tests ====================

// genScanJob generates random ScanJob instances for property testing
func genScanJob() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Name
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Target
		genProfile(),                         // Profile
		genStatus(),                          // Status
		gen.IntRange(0, 100),                 // Progress
		genTaskState(),                       // State
		genConfig(),                          // Config
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // OrgID
	).Map(func(values []interface{}) ScanJob {
		now := time.Now().Truncate(time.Millisecond) // Truncate for JSON compatibility
		return ScanJob{
			ID:       primitive.NewObjectID(),
			TaskID:   "task-" + values[0].(string),
			Name:     values[0].(string),
			Target:   values[1].(string),
			Profile:  values[2].(Profile),
			Status:   values[3].(Status),
			Progress: values[4].(int),
			State:    values[5].(TaskState),
			Config:   values[6].(Config),
			OrgID:    values[7].(string),
			Created:  now,
			Updated:  now,
		}
	})
}

// genScanTarget generates random ScanTarget instances for property testing
func genScanTarget() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // JobID
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Host
		gen.SliceOfN(5, gen.IntRange(1, 65535)), // Ports
		gen.SliceOfN(3, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })), // Services
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Category
		gen.IntRange(1, 10),                  // Priority
	).Map(func(values []interface{}) ScanTarget {
		now := time.Now().Truncate(time.Millisecond)
		return ScanTarget{
			ID:       primitive.NewObjectID(),
			JobID:    values[0].(string),
			Host:     values[1].(string),
			Ports:    values[2].([]int),
			Services: values[3].([]string),
			Category: values[4].(string),
			Priority: values[5].(int),
			Created:  now,
			Updated:  now,
		}
	})
}

// genScanResult generates random ScanResult instances for property testing
func genScanResult() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // JobID
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // TargetID
		gen.SliceOfN(3, genFinding()),        // Findings
		gen.SliceOfN(2, genAsset()),          // Assets
		gen.Float64Range(0.0, 100.0),         // RiskScore
		genRiskLevel(),                       // RiskLevel
	).Map(func(values []interface{}) ScanResult {
		now := time.Now().Truncate(time.Millisecond)
		return ScanResult{
			ID:        primitive.NewObjectID(),
			JobID:     values[0].(string),
			TargetID:  values[1].(string),
			Findings:  values[2].([]Finding),
			Assets:    values[3].([]Asset),
			RiskScore: values[4].(float64),
			RiskLevel: values[5].(string),
			Completed: now,
			Created:   now,
			Updated:   now,
		}
	})
}

// genProfile generates random Profile instances
func genProfile() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.SliceOfN(3, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })),
		gen.MapOf(gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), gen.AlphaString().SuchThat(func(s string) bool { return s != "" })),
	).Map(func(values []interface{}) Profile {
		config := values[4].(map[string]string)
		// Ensure at least one entry
		if len(config) == 0 {
			config["default"] = "value"
		}
		return Profile{
			ID:          values[0].(string),
			Name:        values[1].(string),
			Description: values[2].(string),
			Tools:       values[3].([]string),
			Config:      config,
		}
	})
}

// genStatus generates random Status values
func genStatus() gopter.Gen {
	return gen.OneConstOf(
		StatusCreated,
		StatusPending,
		StatusStarted,
		StatusPaused,
		StatusSuccess,
		StatusFailure,
		StatusRevoked,
		StatusStopped,
	)
}

// genTaskState generates random TaskState instances
func genTaskState() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.MapOf(gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), gen.AlphaString().SuchThat(func(s string) bool { return s != "" })),
		gen.SliceOfN(2, genSubTask()),
	).Map(func(values []interface{}) TaskState {
		// Convert string map to interface{} map
		stringMap := values[1].(map[string]string)
		interfaceMap := make(map[string]interface{})
		for k, v := range stringMap {
			interfaceMap[k] = v
		}
		// Ensure at least one entry
		if len(interfaceMap) == 0 {
			interfaceMap["default"] = "value"
		}
		
		return TaskState{
			Phase:    values[0].(string),
			Data:     interfaceMap,
			SubTasks: values[2].([]SubTask),
		}
	})
}

// genSubTask generates random SubTask instances
func genSubTask() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		genStatus(),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	).Map(func(values []interface{}) SubTask {
		now := time.Now().Truncate(time.Millisecond)
		return SubTask{
			ID:      values[0].(string),
			Name:    values[1].(string),
			Status:  values[2].(Status),
			Worker:  values[3].(string),
			Result:  values[4].(string),
			Created: now,
		}
	})
}

// genConfig generates random Config instances
func genConfig() gopter.Gen {
	return gen.MapOf(gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), gen.AlphaString().SuchThat(func(s string) bool { return s != "" })).Map(func(m map[string]string) Config {
		result := make(Config)
		for k, v := range m {
			result[k] = v
		}
		// Ensure at least one entry
		if len(result) == 0 {
			result["default"] = "value"
		}
		return result
	})
}

// genFinding generates random Finding instances
func genFinding() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })),
		gen.MapOf(gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), gen.AlphaString().SuchThat(func(s string) bool { return s != "" })),
		gen.Float64Range(0.0, 100.0),
	).Map(func(values []interface{}) Finding {
		now := time.Now().Truncate(time.Millisecond)
		metadata := values[7].(map[string]string)
		// Ensure at least one entry
		if len(metadata) == 0 {
			metadata["default"] = "value"
		}
		return Finding{
			ID:          values[0].(string),
			Type:        values[1].(string),
			Severity:    values[2].(string),
			Title:       values[3].(string),
			Description: values[4].(string),
			Evidence:    values[5].(string),
			References:  values[6].([]string),
			Metadata:    metadata,
			RiskScore:   values[8].(float64),
			Discovered:  now,
		}
	})
}

// genAsset generates random Asset instances (simplified for testing)
func genAsset() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Authority
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Host
		gen.IntRange(1, 65535),               // Port
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Category
		genIP(),                              // IP
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Domain
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Service
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Server
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Banner
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Title
		gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })), // App
		gen.SliceOfN(2, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })), // Fingerprints
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // HttpStatus
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // HttpHeader
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // HttpBody
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Cert
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // IconHash
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // IconHashFile
		gen.SliceOfN(10, gen.UInt8()).Map(func(slice []uint8) []byte {
			result := make([]byte, len(slice))
			for i, v := range slice {
				result[i] = byte(v)
			}
			return result
		}),                               // IconHashBytes
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Screenshot
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // OrgId
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // ColorTag
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Memo
		gen.Bool(),                           // IsCDN
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // CName
		gen.Bool(),                           // IsCloud
		gen.Bool(),                           // IsHTTP
		gen.Bool(),                           // IsNewAsset
		gen.Bool(),                           // IsUpdated
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // TaskId
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // LastTaskId
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }), // Source
		gen.Float64Range(0.0, 100.0),         // RiskScore
		genRiskLevel(),                       // RiskLevel
	).Map(func(values []interface{}) Asset {
		now := time.Now().Truncate(time.Millisecond)
		return Asset{
			Id:            primitive.NewObjectID(),
			Authority:     values[0].(string),
			Host:          values[1].(string),
			Port:          values[2].(int),
			Category:      values[3].(string),
			Ip:            values[4].(IP),
			Domain:        values[5].(string),
			Service:       values[6].(string),
			Server:        values[7].(string),
			Banner:        values[8].(string),
			Title:         values[9].(string),
			App:           values[10].([]string),
			Fingerprints:  values[11].([]string),
			HttpStatus:    values[12].(string),
			HttpHeader:    values[13].(string),
			HttpBody:      values[14].(string),
			Cert:          values[15].(string),
			IconHash:      values[16].(string),
			IconHashFile:  values[17].(string),
			IconHashBytes: values[18].([]byte),
			Screenshot:    values[19].(string),
			OrgId:         values[20].(string),
			ColorTag:      values[21].(string),
			Memo:          values[22].(string),
			IsCDN:         values[23].(bool),
			CName:         values[24].(string),
			IsCloud:       values[25].(bool),
			IsHTTP:        values[26].(bool),
			IsNewAsset:    values[27].(bool),
			IsUpdated:     values[28].(bool),
			TaskId:        "task-" + values[0].(string),
			LastTaskId:    values[30].(string),
			Source:        values[31].(string),
			CreateTime:    now,
			UpdateTime:    now,
			RiskScore:     values[32].(float64),
			RiskLevel:     values[33].(string),
		}
	})
}

// genIP generates random IP instances
func genIP() gopter.Gen {
	return gopter.CombineGens(
		gen.SliceOfN(2, genIPV4()),
		gen.SliceOfN(2, genIPV6()),
	).Map(func(values []interface{}) IP {
		return IP{
			IpV4: values[0].([]IPV4),
			IpV6: values[1].([]IPV6),
		}
	})
}

// genIPV4 generates random IPV4 instances
func genIPV4() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.UInt32(),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	).Map(func(values []interface{}) IPV4 {
		return IPV4{
			IPName:   values[0].(string),
			IPInt:    values[1].(uint32),
			Location: values[2].(string),
		}
	})
}

// genIPV6 generates random IPV6 instances
func genIPV6() gopter.Gen {
	return gopter.CombineGens(
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
		gen.AlphaString().SuchThat(func(s string) bool { return s != "" }),
	).Map(func(values []interface{}) IPV6 {
		return IPV6{
			IPName:   values[0].(string),
			Location: values[1].(string),
		}
	})
}

// genRiskLevel generates random risk level values
func genRiskLevel() gopter.Gen {
	return gen.OneConstOf("critical", "high", "medium", "low", "info", "unknown")
}

// ==================== Property Tests ====================

// TestProperty1_DataStructureSerializationRoundTrip tests Property 1: Data Structure Serialization Round Trip
// **Property 1: Data Structure Serialization Round Trip**
// **Validates: Requirements 1.5**
// For any core data structure (ScanJob, ScanTarget, ScanResult), serializing then deserializing should produce an equivalent object
func TestProperty1_DataStructureSerializationRoundTrip(t *testing.T) {
	// Feature: cscan-refactoring, Property 1: Data Structure Serialization Round Trip
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: ScanJob JSON serialization round trip
	properties.Property("ScanJob JSON serialization round trip", prop.ForAll(
		func(job ScanJob) bool {
			// Serialize to JSON
			jsonData, err := json.Marshal(job)
			if err != nil {
				return false
			}

			// Deserialize from JSON
			var decoded ScanJob
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				return false
			}

			// Compare essential fields (excluding ObjectID which may differ in representation)
			return job.TaskID == decoded.TaskID &&
				job.Name == decoded.Name &&
				job.Target == decoded.Target &&
				job.Status == decoded.Status &&
				job.Progress == decoded.Progress &&
				job.OrgID == decoded.OrgID &&
				job.Created.Equal(decoded.Created) &&
				job.Updated.Equal(decoded.Updated)
		},
		genScanJob(),
	))

	// Property: ScanJob BSON serialization round trip
	properties.Property("ScanJob BSON serialization round trip", prop.ForAll(
		func(job ScanJob) bool {
			// Serialize to BSON
			bsonData, err := bson.Marshal(job)
			if err != nil {
				return false
			}

			// Deserialize from BSON
			var decoded ScanJob
			err = bson.Unmarshal(bsonData, &decoded)
			if err != nil {
				return false
			}

			// Compare essential fields (BSON may omit empty fields with omitempty tags)
			return job.TaskID == decoded.TaskID &&
				job.Name == decoded.Name &&
				job.Target == decoded.Target &&
				job.Status == decoded.Status &&
				job.Progress == decoded.Progress &&
				job.OrgID == decoded.OrgID &&
				job.Created.Equal(decoded.Created) &&
				job.Updated.Equal(decoded.Updated) &&
				timePointersEqual(job.Started, decoded.Started) &&
				timePointersEqual(job.Ended, decoded.Ended) &&
				profilesEqual(job.Profile, decoded.Profile) &&
				taskStatesEqual(job.State, decoded.State) &&
				configsEqual(job.Config, decoded.Config)
		},
		genScanJob(),
	))

	// Property: ScanTarget JSON serialization round trip
	properties.Property("ScanTarget JSON serialization round trip", prop.ForAll(
		func(target ScanTarget) bool {
			// Serialize to JSON
			jsonData, err := json.Marshal(target)
			if err != nil {
				return false
			}

			// Deserialize from JSON
			var decoded ScanTarget
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				return false
			}

			// Compare essential fields
			return target.JobID == decoded.JobID &&
				target.Host == decoded.Host &&
				reflect.DeepEqual(target.Ports, decoded.Ports) &&
				reflect.DeepEqual(target.Services, decoded.Services) &&
				target.Category == decoded.Category &&
				target.Priority == decoded.Priority &&
				target.Created.Equal(decoded.Created) &&
				target.Updated.Equal(decoded.Updated)
		},
		genScanTarget(),
	))

	// Property: ScanTarget BSON serialization round trip
	properties.Property("ScanTarget BSON serialization round trip", prop.ForAll(
		func(target ScanTarget) bool {
			// Serialize to BSON
			bsonData, err := bson.Marshal(target)
			if err != nil {
				return false
			}

			// Deserialize from BSON
			var decoded ScanTarget
			err = bson.Unmarshal(bsonData, &decoded)
			if err != nil {
				return false
			}

			// Compare essential fields (BSON handles slices and basic types consistently)
			return target.JobID == decoded.JobID &&
				target.Host == decoded.Host &&
				reflect.DeepEqual(target.Ports, decoded.Ports) &&
				reflect.DeepEqual(target.Services, decoded.Services) &&
				target.Category == decoded.Category &&
				target.Priority == decoded.Priority &&
				target.Created.Equal(decoded.Created) &&
				target.Updated.Equal(decoded.Updated)
		},
		genScanTarget(),
	))

	// Property: ScanResult JSON serialization round trip
	properties.Property("ScanResult JSON serialization round trip", prop.ForAll(
		func(result ScanResult) bool {
			// Serialize to JSON
			jsonData, err := json.Marshal(result)
			if err != nil {
				return false
			}

			// Deserialize from JSON
			var decoded ScanResult
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				return false
			}

			// Compare essential fields
			return result.JobID == decoded.JobID &&
				result.TargetID == decoded.TargetID &&
				result.RiskScore == decoded.RiskScore &&
				result.RiskLevel == decoded.RiskLevel &&
				result.Completed.Equal(decoded.Completed) &&
				result.Created.Equal(decoded.Created) &&
				result.Updated.Equal(decoded.Updated)
		},
		genScanResult(),
	))

	// Property: ScanResult BSON serialization round trip
	properties.Property("ScanResult BSON serialization round trip", prop.ForAll(
		func(result ScanResult) bool {
			// Serialize to BSON
			bsonData, err := bson.Marshal(result)
			if err != nil {
				return false
			}

			// Deserialize from BSON
			var decoded ScanResult
			err = bson.Unmarshal(bsonData, &decoded)
			if err != nil {
				return false
			}

			// Compare essential fields (BSON handles complex nested structures)
			return result.JobID == decoded.JobID &&
				result.TargetID == decoded.TargetID &&
				result.RiskScore == decoded.RiskScore &&
				result.RiskLevel == decoded.RiskLevel &&
				result.Completed.Equal(decoded.Completed) &&
				result.Created.Equal(decoded.Created) &&
				result.Updated.Equal(decoded.Updated) &&
				findingsEqual(result.Findings, decoded.Findings) &&
				assetsEqual(result.Assets, decoded.Assets)
		},
		genScanResult(),
	))

	// Property: Profile serialization round trip
	properties.Property("Profile serialization round trip", prop.ForAll(
		func(profile Profile) bool {
			// JSON round trip
			jsonData, err := json.Marshal(profile)
			if err != nil {
				return false
			}

			var jsonDecoded Profile
			err = json.Unmarshal(jsonData, &jsonDecoded)
			if err != nil {
				return false
			}

			// BSON round trip
			bsonData, err := bson.Marshal(profile)
			if err != nil {
				return false
			}

			var bsonDecoded Profile
			err = bson.Unmarshal(bsonData, &bsonDecoded)
			if err != nil {
				return false
			}

			// Both should be equal to original (use semantic equality for BSON)
			jsonMatch := reflect.DeepEqual(profile, jsonDecoded)
			bsonMatch := profilesEqual(profile, bsonDecoded)
			return jsonMatch && bsonMatch
		},
		genProfile(),
	))

	// Property: Finding serialization round trip
	properties.Property("Finding serialization round trip", prop.ForAll(
		func(finding Finding) bool {
			// JSON round trip
			jsonData, err := json.Marshal(finding)
			if err != nil {
				return false
			}

			var jsonDecoded Finding
			err = json.Unmarshal(jsonData, &jsonDecoded)
			if err != nil {
				return false
			}

			// BSON round trip
			bsonData, err := bson.Marshal(finding)
			if err != nil {
				return false
			}

			var bsonDecoded Finding
			err = bson.Unmarshal(bsonData, &bsonDecoded)
			if err != nil {
				return false
			}

			// Compare essential fields (time comparison needs special handling)
			jsonMatch := finding.ID == jsonDecoded.ID &&
				finding.Type == jsonDecoded.Type &&
				finding.Severity == jsonDecoded.Severity &&
				finding.Title == jsonDecoded.Title &&
				finding.RiskScore == jsonDecoded.RiskScore &&
				finding.Discovered.Equal(jsonDecoded.Discovered)

			bsonMatch := findingEqual(finding, bsonDecoded)

			return jsonMatch && bsonMatch
		},
		genFinding(),
	))

	// Property: Config serialization round trip
	properties.Property("Config serialization round trip", prop.ForAll(
		func(config Config) bool {
			// JSON round trip
			jsonData, err := json.Marshal(config)
			if err != nil {
				return false
			}

			var jsonDecoded Config
			err = json.Unmarshal(jsonData, &jsonDecoded)
			if err != nil {
				return false
			}

			// BSON round trip
			bsonData, err := bson.Marshal(config)
			if err != nil {
				return false
			}

			var bsonDecoded Config
			err = bson.Unmarshal(bsonData, &bsonDecoded)
			if err != nil {
				return false
			}

			// Both should be equal to original (use semantic equality for BSON)
			jsonMatch := reflect.DeepEqual(config, jsonDecoded)
			bsonMatch := configsEqual(config, bsonDecoded)
			return jsonMatch && bsonMatch
		},
		genConfig(),
	))

	// Property: TaskState serialization round trip
	properties.Property("TaskState serialization round trip", prop.ForAll(
		func(state TaskState) bool {
			// JSON round trip
			jsonData, err := json.Marshal(state)
			if err != nil {
				return false
			}

			var jsonDecoded TaskState
			err = json.Unmarshal(jsonData, &jsonDecoded)
			if err != nil {
				return false
			}

			// BSON round trip
			bsonData, err := bson.Marshal(state)
			if err != nil {
				return false
			}

			var bsonDecoded TaskState
			err = bson.Unmarshal(bsonData, &bsonDecoded)
			if err != nil {
				return false
			}

			// Compare essential fields
			jsonMatch := state.Phase == jsonDecoded.Phase &&
				len(state.SubTasks) == len(jsonDecoded.SubTasks)

			bsonMatch := taskStatesEqual(state, bsonDecoded)

			return jsonMatch && bsonMatch
		},
		genTaskState(),
	))

	properties.TestingRun(t)
}

// ==================== Helper Functions for BSON Semantic Equality ====================

// timePointersEqual compares two time pointers, handling nil cases
func timePointersEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

// profilesEqual compares two Profile structs semantically
func profilesEqual(a, b Profile) bool {
	return a.ID == b.ID &&
		a.Name == b.Name &&
		a.Description == b.Description &&
		reflect.DeepEqual(a.Tools, b.Tools) &&
		reflect.DeepEqual(a.Config, b.Config)
}

// taskStatesEqual compares two TaskState structs semantically
func taskStatesEqual(a, b TaskState) bool {
	if a.Phase != b.Phase {
		return false
	}
	if !reflect.DeepEqual(a.Data, b.Data) {
		return false
	}
	if len(a.SubTasks) != len(b.SubTasks) {
		return false
	}
	for i, subTask := range a.SubTasks {
		if !subTasksEqual(subTask, b.SubTasks[i]) {
			return false
		}
	}
	return timePointersEqual(a.CompletedAt, b.CompletedAt)
}

// subTasksEqual compares two SubTask structs semantically
func subTasksEqual(a, b SubTask) bool {
	return a.ID == b.ID &&
		a.Name == b.Name &&
		a.Status == b.Status &&
		a.Worker == b.Worker &&
		a.Result == b.Result &&
		a.Created.Equal(b.Created) &&
		timePointersEqual(a.Started, b.Started) &&
		timePointersEqual(a.Ended, b.Ended)
}

// configsEqual compares two Config maps semantically
func configsEqual(a, b Config) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, exists := b[k]; !exists || !reflect.DeepEqual(v, bv) {
			return false
		}
	}
	return true
}

// findingsEqual compares two Finding slices semantically
func findingsEqual(a, b []Finding) bool {
	if len(a) != len(b) {
		return false
	}
	for i, finding := range a {
		if !findingEqual(finding, b[i]) {
			return false
		}
	}
	return true
}

// findingEqual compares two Finding structs semantically
func findingEqual(a, b Finding) bool {
	return a.ID == b.ID &&
		a.Type == b.Type &&
		a.Severity == b.Severity &&
		a.Title == b.Title &&
		a.Description == b.Description &&
		a.Evidence == b.Evidence &&
		reflect.DeepEqual(a.References, b.References) &&
		reflect.DeepEqual(a.Metadata, b.Metadata) &&
		a.RiskScore == b.RiskScore &&
		a.Discovered.Equal(b.Discovered)
}

// assetsEqual compares two Asset slices semantically
func assetsEqual(a, b []Asset) bool {
	if len(a) != len(b) {
		return false
	}
	for i, asset := range a {
		if !assetEqual(asset, b[i]) {
			return false
		}
	}
	return true
}

// assetEqual compares two Asset structs semantically
func assetEqual(a, b Asset) bool {
	return a.Authority == b.Authority &&
		a.Host == b.Host &&
		a.Port == b.Port &&
		a.Category == b.Category &&
		ipEqual(a.Ip, b.Ip) &&
		a.Domain == b.Domain &&
		a.Service == b.Service &&
		a.Server == b.Server &&
		a.Banner == b.Banner &&
		a.Title == b.Title &&
		reflect.DeepEqual(a.App, b.App) &&
		reflect.DeepEqual(a.Fingerprints, b.Fingerprints) &&
		a.HttpStatus == b.HttpStatus &&
		a.HttpHeader == b.HttpHeader &&
		a.HttpBody == b.HttpBody &&
		a.Cert == b.Cert &&
		a.IconHash == b.IconHash &&
		a.IconHashFile == b.IconHashFile &&
		reflect.DeepEqual(a.IconHashBytes, b.IconHashBytes) &&
		a.Screenshot == b.Screenshot &&
		a.OrgId == b.OrgId &&
		a.ColorTag == b.ColorTag &&
		a.Memo == b.Memo &&
		a.IsCDN == b.IsCDN &&
		a.CName == b.CName &&
		a.IsCloud == b.IsCloud &&
		a.IsHTTP == b.IsHTTP &&
		a.IsNewAsset == b.IsNewAsset &&
		a.IsUpdated == b.IsUpdated &&
		a.TaskId == b.TaskId &&
		a.LastTaskId == b.LastTaskId &&
		a.Source == b.Source &&
		a.CreateTime.Equal(b.CreateTime) &&
		a.UpdateTime.Equal(b.UpdateTime) &&
		a.RiskScore == b.RiskScore &&
		a.RiskLevel == b.RiskLevel
}

// ipEqual compares two IP structs semantically
func ipEqual(a, b IP) bool {
	return reflect.DeepEqual(a.IpV4, b.IpV4) &&
		reflect.DeepEqual(a.IpV6, b.IpV6)
}

// ==================== Unit Tests ====================

// TestScanJobValidation tests ScanJob validation
func TestScanJobValidation(t *testing.T) {
	tests := []struct {
		name    string
		job     ScanJob
		wantErr bool
	}{
		{
			name: "valid job",
			job: ScanJob{
				Name:   "test-job",
				Target: "example.com",
				Profile: Profile{
					ID:   "profile-1",
					Name: "Basic Scan",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			job: ScanJob{
				Target: "example.com",
				Profile: Profile{
					ID:   "profile-1",
					Name: "Basic Scan",
				},
			},
			wantErr: true,
		},
		{
			name: "missing target",
			job: ScanJob{
				Name: "test-job",
				Profile: Profile{
					ID:   "profile-1",
					Name: "Basic Scan",
				},
			},
			wantErr: true,
		},
		{
			name: "missing profile",
			job: ScanJob{
				Name:   "test-job",
				Target: "example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.job.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanJob.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScanTargetValidation tests ScanTarget validation
func TestScanTargetValidation(t *testing.T) {
	tests := []struct {
		name    string
		target  ScanTarget
		wantErr bool
	}{
		{
			name: "valid target",
			target: ScanTarget{
				JobID: "job-1",
				Host:  "example.com",
			},
			wantErr: false,
		},
		{
			name: "missing jobId",
			target: ScanTarget{
				Host: "example.com",
			},
			wantErr: true,
		},
		{
			name: "missing host",
			target: ScanTarget{
				JobID: "job-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanTarget.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestScanResultValidation tests ScanResult validation
func TestScanResultValidation(t *testing.T) {
	tests := []struct {
		name    string
		result  ScanResult
		wantErr bool
	}{
		{
			name: "valid result",
			result: ScanResult{
				JobID:    "job-1",
				TargetID: "target-1",
			},
			wantErr: false,
		},
		{
			name: "missing jobId",
			result: ScanResult{
				TargetID: "target-1",
			},
			wantErr: true,
		},
		{
			name: "missing targetId",
			result: ScanResult{
				JobID: "job-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanResult.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}