package crossplane

import (
	"fmt"

	"github.com/riptano/iac_generator_cli/internal/utils"
	"github.com/riptano/iac_generator_cli/pkg/models"
	"gopkg.in/yaml.v3"
)

// K8sObject represents a generic Kubernetes object structure
type K8sObject struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   Metadata               `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
}

// Metadata represents Kubernetes resource metadata
type Metadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// YAML returns the YAML representation of the K8sObject
func (obj K8sObject) YAML() string {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("Error marshaling YAML: %v", err)
	}
	return string(data)
}

// AddMetadataField adds a field to the metadata section
func (obj *K8sObject) AddMetadataField(key string, value string) {
	switch key {
	case "name":
		obj.Metadata.Name = value
	case "namespace":
		obj.Metadata.Namespace = value
	default:
		// For other fields, add to labels
		if obj.Metadata.Labels == nil {
			obj.Metadata.Labels = make(map[string]string)
		}
		obj.Metadata.Labels[key] = value
	}
}

// AddMetadataAnnotation adds an annotation to the metadata
func (obj *K8sObject) AddMetadataAnnotation(key string, value string) {
	if obj.Metadata.Annotations == nil {
		obj.Metadata.Annotations = make(map[string]string)
	}
	obj.Metadata.Annotations[key] = value
}

// AddField adds a field at the top level of the object
func (obj *K8sObject) AddField(key string, value interface{}) {
	// Handle special fields
	switch key {
	case "apiVersion":
		if strValue, ok := value.(string); ok {
			obj.APIVersion = strValue
		}
	case "kind":
		if strValue, ok := value.(string); ok {
			obj.Kind = strValue
		}
	case "metadata":
		// Intentionally left empty, use AddMetadataField instead
	case "spec":
		if mapValue, ok := value.(map[string]interface{}); ok {
			obj.Spec = mapValue
		}
	default:
		// For custom fields, add directly
		// This is for non-standard fields like 'data' in Secret objects
		if obj.Spec == nil {
			obj.Spec = make(map[string]interface{})
		}
		obj.Spec[key] = value
	}
}

// GenerateYAML converts a K8sObject to YAML
func GenerateYAML(obj K8sObject) (string, error) {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object to YAML: %w", err)
	}
	return string(data), nil
}

// WriteYAML writes a K8sObject to a YAML file
func WriteYAML(obj K8sObject, filePath string) error {
	yamlStr, err := GenerateYAML(obj)
	if err != nil {
		return err
	}
	return utils.WriteToFile(filePath, yamlStr)
}

// WriteMultiYAML writes multiple K8sObjects to a YAML file with separators
func WriteMultiYAML(objects []K8sObject, filePath string) error {
	var content string
	for i, obj := range objects {
		yamlStr, err := GenerateYAML(obj)
		if err != nil {
			return err
		}
		content += yamlStr
		if i < len(objects)-1 {
			content += "\n---\n"
		}
	}
	return utils.WriteToFile(filePath, content)
}

// NewK8sObject creates a new Kubernetes object with the specified API version, kind, and name
func NewK8sObject(apiVersion, kind, name string) K8sObject {
	return K8sObject{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: Metadata{
			Name: name,
		},
		Spec: make(map[string]interface{}),
	}
}

// SetNamespace sets the namespace for a Kubernetes object
func (obj *K8sObject) SetNamespace(namespace string) {
	obj.Metadata.Namespace = namespace
}

// AddLabel adds a label to a Kubernetes object
func (obj *K8sObject) AddLabel(key, value string) {
	if obj.Metadata.Labels == nil {
		obj.Metadata.Labels = make(map[string]string)
	}
	obj.Metadata.Labels[key] = value
}

// AddAnnotation adds an annotation to a Kubernetes object
func (obj *K8sObject) AddAnnotation(key, value string) {
	if obj.Metadata.Annotations == nil {
		obj.Metadata.Annotations = make(map[string]string)
	}
	obj.Metadata.Annotations[key] = value
}

// SetSpecField sets a field in the spec of a Kubernetes object
func (obj *K8sObject) SetSpecField(key string, value interface{}) {
	obj.Spec[key] = value
}

// AddNestedSpecField adds a nested field in the spec of a Kubernetes object
func (obj *K8sObject) AddNestedSpecField(path []string, value interface{}) {
	if len(path) == 0 {
		return
	}

	current := obj.Spec
	for i, key := range path {
		if i == len(path)-1 {
			// Last key, set the value
			current[key] = value
			return
		}

		// Not the last key, need to navigate or create the map
		if _, ok := current[key]; !ok {
			current[key] = make(map[string]interface{})
		}

		if nestedMap, ok := current[key].(map[string]interface{}); ok {
			current = nestedMap
		} else {
			// If it's not a map, convert it to one
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}
}

// ConvertResourceToK8sObject converts an internal resource model to a Crossplane K8s object
func ConvertResourceToK8sObject(resource models.Resource) (K8sObject, error) {
	apiVersion, kind, err := mapResourceTypeToK8s(resource.Type)
	if err != nil {
		return K8sObject{}, err
	}

	obj := NewK8sObject(apiVersion, kind, resource.Name)
	
	// Add common labels
	obj.AddLabel("app.kubernetes.io/part-of", "infrastructure")
	obj.AddLabel("app.kubernetes.io/managed-by", "crossplane")
	
	// Convert properties
	for _, prop := range resource.Properties {
		// Handle special properties like dependencies
		if prop.Name == "dependsOn" {
			// Skip this for now, will handle dependencies separately
			continue
		}
		
		// Map the property name to the Crossplane format
		crossplanePropName := mapPropertyName(prop.Name)
		obj.AddNestedSpecField([]string{"forProvider", crossplanePropName}, prop.Value)
	}
	
	// Add provider config reference
	obj.AddNestedSpecField([]string{"providerConfigRef", "name"}, "aws-provider")
	
	// Handle dependencies if they exist
	if len(resource.DependsOn) > 0 {
		refs := make([]map[string]interface{}, 0, len(resource.DependsOn))
		for _, dep := range resource.DependsOn {
			refs = append(refs, map[string]interface{}{
				"name": dep,
			})
		}
		obj.AddNestedSpecField([]string{"dependsOn"}, refs)
	}
	
	return obj, nil
}

// mapResourceTypeToK8s maps internal resource types to Crossplane API/Kind
func mapResourceTypeToK8s(resourceType models.ResourceType) (string, string, error) {
	mapping := map[models.ResourceType]struct {
		APIVersion string
		Kind       string
	}{
		models.ResourceVPC: {
			APIVersion: "ec2.aws.crossplane.io/v1beta1",
			Kind:       "VPC",
		},
		models.ResourceSubnet: {
			APIVersion: "ec2.aws.crossplane.io/v1beta1",
			Kind:       "Subnet",
		},
		models.ResourceIGW: {
			APIVersion: "ec2.aws.crossplane.io/v1beta1",
			Kind:       "InternetGateway",
		},
		models.ResourceNATGateway: {
			APIVersion: "ec2.aws.crossplane.io/v1beta1",
			Kind:       "NATGateway",
		},
		models.ResourceSecurityGroup: {
			APIVersion: "ec2.aws.crossplane.io/v1beta1",
			Kind:       "SecurityGroup",
		},
		models.ResourceIAMRole: {
			APIVersion: "iam.aws.crossplane.io/v1beta1",
			Kind:       "Role",
		},
		models.ResourceEKSCluster: {
			APIVersion: "eks.aws.crossplane.io/v1beta1",
			Kind:       "Cluster",
		},
		models.ResourceNodeGroup: {
			APIVersion: "eks.aws.crossplane.io/v1beta1",
			Kind:       "NodeGroup",
		},
		models.ResourceS3Bucket: {
			APIVersion: "s3.aws.crossplane.io/v1beta1",
			Kind:       "Bucket",
		},
		models.ResourceEC2Instance: {
			APIVersion: "ec2.aws.crossplane.io/v1beta1",
			Kind:       "Instance",
		},
	}

	if mapping, ok := mapping[resourceType]; ok {
		return mapping.APIVersion, mapping.Kind, nil
	}

	return "", "", fmt.Errorf("unsupported resource type for Crossplane: %s", resourceType)
}

// mapPropertyName maps internal property names to Crossplane property names
func mapPropertyName(propName string) string {
	mapping := map[string]string{
		"cidr_block":           "cidrBlock",
		"enable_dns_support":   "enableDnsSupport",
		"enable_dns_hostnames": "enableDnsHostnames",
		"availability_zone":    "availabilityZone",
		"vpc_id":               "vpcId",
		"allocation_id":        "allocationId",
		"subnet_id":            "subnetId",
		"node_role":            "nodeRole",
		"instance_type":        "instanceType",
		"node_count":           "desiredSize",
		"version":              "version",
		"subnet_ids":           "subnetIds",
		"role_arn":             "roleArn",
		"endpoint_public_access": "endpointPublicAccess",
		"endpoint_private_access": "endpointPrivateAccess",
	}

	if mapped, ok := mapping[propName]; ok {
		return mapped
	}

	// If no mapping exists, return the original name
	return propName
}