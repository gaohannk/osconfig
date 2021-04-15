//  Copyright 2021 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ospolicies

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/osconfig/e2e_tests/utils"
	"google.golang.org/protobuf/types/known/durationpb"

	osconfig "github.com/GoogleCloudPlatform/osconfig/e2e_tests/internal/google.golang.org/genproto/googleapis/cloud/osconfig/v1alpha"
	osconfigpb "github.com/GoogleCloudPlatform/osconfig/e2e_tests/internal/google.golang.org/genproto/googleapis/cloud/osconfig/v1alpha"
)

const (
	packageInstalled    = "osconfig_tests/pkg_installed"
	packageNotInstalled = "osconfig_tests/pkg_not_installed"
	fileExists          = "osconfig_tests/file_exists"
	fileDNE             = "osconfig_tests/file_does_not_exist"
	osconfigTestRepo    = "osconfig-agent-test-repository"
	testResourceBucket  = "osconfig-agent-end2end-test-resources"
	yumTestRepoBaseURL  = "https://packages.cloud.google.com/yum/repos/osconfig-agent-test-repository"
	aptTestRepoBaseURL  = "http://packages.cloud.google.com/apt"
	gooTestRepoURL      = "https://packages.cloud.google.com/yuck/repos/osconfig-agent-test-repository"
	aptRaptureGpgKey    = "https://packages.cloud.google.com/apt/doc/apt-key.gpg"
)

var (
	yumRaptureGpgKeys = []string{"https://packages.cloud.google.com/yum/doc/yum-key.gpg", "https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg"}
)

var wantPackageCompliances = []*osconfigpb.OSPolicyResourceCompliance{
	{
		OsPolicyResourceId: "install-package",
		ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
		},
		State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
	},
	{
		OsPolicyResourceId: "remove-package",
		ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
		},
		State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
	},
}

var wantRepositoryCompliances = []*osconfigpb.OSPolicyResourceCompliance{
	{
		OsPolicyResourceId: "install-repo",
		ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
		},
		State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
	},
	{
		OsPolicyResourceId: "install-package",
		ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
			{
				Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
				Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
			},
		},
		State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
	},
}

func buildAptTestSetup(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	testName := packageResourceApt
	machineType := "e2-medium"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Apt{
											Apt: &osconfigpb.OSPolicy_Resource_PackageResource_APT{Name: "ed"},
										},
									},
								},
							},
							{
								Id: "remove-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_REMOVED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Apt{
											Apt: &osconfigpb.OSPolicy_Resource_PackageResource_APT{Name: "vim"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantPackageCompliances,
		},
	}
	ss := getStartupScriptPackage(name, "apt")
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled, packageNotInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildYumTestSetup(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	testName := packageResourceYum
	machineType := "e2-medium"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Yum{
											Yum: &osconfigpb.OSPolicy_Resource_PackageResource_YUM{Name: "ed"},
										},
									},
								},
							},
							{
								Id: "remove-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_REMOVED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Yum{
											Yum: &osconfigpb.OSPolicy_Resource_PackageResource_YUM{Name: "nano"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantPackageCompliances,
		},
	}
	ss := getStartupScriptPackage(name, "yum")
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled, packageNotInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildZypperTestSetup(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	testName := packageResourceZypper
	machineType := "e2-medium"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Zypper_{
											Zypper: &osconfigpb.OSPolicy_Resource_PackageResource_Zypper{Name: "ed"},
										},
									},
								},
							},
							{
								Id: "remove-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_REMOVED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Zypper_{
											Zypper: &osconfigpb.OSPolicy_Resource_PackageResource_Zypper{Name: "vim"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantPackageCompliances,
		},
	}
	ss := getStartupScriptPackage(name, "zypper")
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled, packageNotInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildGooGetTestSetup(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	testName := packageResourceGoo
	machineType := "e2-standard-2"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Googet{
											Googet: &osconfigpb.OSPolicy_Resource_PackageResource_GooGet{Name: "cowsay"},
										},
									},
								},
							},
							{
								Id: "remove-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_REMOVED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Googet{
											Googet: &osconfigpb.OSPolicy_Resource_PackageResource_GooGet{Name: "certgen"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantPackageCompliances,
		},
	}
	ss := getStartupScriptPackage(name, "googet")
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled, packageNotInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func addPackageResourceTests(key string) []*osPolicyTestSetup {
	var pkgTestSetup []*osPolicyTestSetup
	for name, image := range utils.HeadAptImages {
		pkgTestSetup = append(pkgTestSetup, buildAptTestSetup(name, image, key))
	}
	for name, image := range utils.HeadELImages {
		pkgTestSetup = append(pkgTestSetup, buildYumTestSetup(name, image, key))
	}
	for name, image := range utils.HeadSUSEImages {
		pkgTestSetup = append(pkgTestSetup, buildZypperTestSetup(name, image, key))
	}
	for name, image := range utils.HeadWindowsImages {
		pkgTestSetup = append(pkgTestSetup, buildGooGetTestSetup(name, image, key))
	}
	return pkgTestSetup
}

func buildAptRepositoryResourceTest(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	packageName := "osconfig-agent-test"
	testName := repositoryResourceApt
	machineType := "e2-medium"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-repo",
								ResourceType: &osconfigpb.OSPolicy_Resource_Repository{
									Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource{
										Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource_Apt{
											Apt: &osconfigpb.OSPolicy_Resource_RepositoryResource_AptRepository{
												ArchiveType:  osconfigpb.OSPolicy_Resource_RepositoryResource_AptRepository_DEB,
												Uri:          aptTestRepoBaseURL,
												Distribution: osconfigTestRepo,
												Components:   []string{"main"},
												GpgKey:       aptRaptureGpgKey,
											},
										},
									},
								},
							},
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Apt{
											Apt: &osconfigpb.OSPolicy_Resource_PackageResource_APT{Name: packageName},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantRepositoryCompliances,
		},
	}
	ss := getStartupScriptRepo(name, "apt", packageName)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildYumRepositoryResourceTest(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	packageName := "osconfig-agent-test"
	testName := repositoryResourceYum
	machineType := "e2-medium"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-repo",
								ResourceType: &osconfigpb.OSPolicy_Resource_Repository{
									Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource{
										Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource_Yum{
											Yum: &osconfigpb.OSPolicy_Resource_RepositoryResource_YumRepository{
												Id:          osconfigTestRepo,
												DisplayName: "Google OSConfig Agent Test Repository",
												BaseUrl:     yumTestRepoBaseURL,
												GpgKeys:     yumRaptureGpgKeys,
											},
										},
									},
								},
							},
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Yum{
											Yum: &osconfigpb.OSPolicy_Resource_PackageResource_YUM{Name: packageName},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantRepositoryCompliances,
		},
	}
	ss := getStartupScriptRepo(name, "yum", packageName)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildZypperRepositoryResourceTest(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	packageName := "osconfig-agent-test"
	testName := repositoryResourceZypper
	machineType := "e2-medium"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-repo",
								ResourceType: &osconfigpb.OSPolicy_Resource_Repository{
									Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource{
										Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource_Zypper{
											Zypper: &osconfigpb.OSPolicy_Resource_RepositoryResource_ZypperRepository{
												Id:          osconfigTestRepo,
												DisplayName: "Google OSConfig Agent Test Repository",
												BaseUrl:     yumTestRepoBaseURL,
												GpgKeys:     yumRaptureGpgKeys,
											},
										},
									},
								},
							},
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Zypper_{
											Zypper: &osconfigpb.OSPolicy_Resource_PackageResource_Zypper{Name: packageName},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantRepositoryCompliances,
		},
	}
	ss := getStartupScriptRepo(name, "yum", packageName)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildGoogetRepositoryResourceTest(name, image, key string) *osPolicyTestSetup {
	assertTimeout := 120 * time.Second
	packageName := "osconfig-agent-test"
	testName := repositoryResourceGoo
	machineType := "e2-standard-2"

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "install-repo",
								ResourceType: &osconfigpb.OSPolicy_Resource_Repository{
									Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource{
										Repository: &osconfigpb.OSPolicy_Resource_RepositoryResource_Goo{
											Goo: &osconfigpb.OSPolicy_Resource_RepositoryResource_GooRepository{
												Name: "Google OSConfig Agent Test Repository",
												Url:  gooTestRepoURL,
											},
										},
									},
								},
							},
							{
								Id: "install-package",
								ResourceType: &osconfigpb.OSPolicy_Resource_Pkg{
									Pkg: &osconfigpb.OSPolicy_Resource_PackageResource{
										DesiredState: osconfigpb.OSPolicy_Resource_PackageResource_INSTALLED,
										SystemPackage: &osconfigpb.OSPolicy_Resource_PackageResource_Googet{
											Googet: &osconfigpb.OSPolicy_Resource_PackageResource_GooGet{Name: packageName},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId:                  testName,
			State:                       osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: wantRepositoryCompliances,
		},
	}
	ss := getStartupScriptRepo(name, "googet", packageName)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{packageInstalled}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func addRepositoryResourceTests(key string) []*osPolicyTestSetup {
	var pkgTestSetup []*osPolicyTestSetup
	for name, image := range utils.HeadAptImages {
		pkgTestSetup = append(pkgTestSetup, buildAptRepositoryResourceTest(name, image, key))
	}
	for name, image := range utils.HeadELImages {
		pkgTestSetup = append(pkgTestSetup, buildYumRepositoryResourceTest(name, image, key))
	}
	for name, image := range utils.HeadSUSEImages {
		pkgTestSetup = append(pkgTestSetup, buildZypperRepositoryResourceTest(name, image, key))
	}
	for name, image := range utils.HeadWindowsImages {
		pkgTestSetup = append(pkgTestSetup, buildGoogetRepositoryResourceTest(name, image, key))
	}
	return pkgTestSetup
}

func buildFileResourceTests(name, image, pkgManager, key string) *osPolicyTestSetup {
	assertTimeout := 180 * time.Second
	testName := fileResource
	machineType := "e2-medium"
	if strings.Contains(image, "windows") {
		machineType = "e2-standard-2"
	}
	dnePath := "/enforce_absent"
	wantPaths := []string{"/from_content", "/from_gcs", "/from_uri", "/from_local"}

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   "files-absent",
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "file-absent",
								ResourceType: &osconfigpb.OSPolicy_Resource_File_{
									File: &osconfigpb.OSPolicy_Resource_FileResource{
										State: osconfigpb.OSPolicy_Resource_FileResource_ABSENT,
										Path:  dnePath,
									},
								},
							},
						},
					},
				},
			},
			{
				Id:   "files-present",
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "file-present-from-content",
								ResourceType: &osconfigpb.OSPolicy_Resource_File_{
									File: &osconfigpb.OSPolicy_Resource_FileResource{
										State:  osconfigpb.OSPolicy_Resource_FileResource_PRESENT,
										Path:   wantPaths[0],
										Source: &osconfigpb.OSPolicy_Resource_FileResource_Content{Content: "something"},
									},
								},
							},
							{
								Id: "file-present-from-gcs",
								ResourceType: &osconfigpb.OSPolicy_Resource_File_{
									File: &osconfigpb.OSPolicy_Resource_FileResource{
										State: osconfigpb.OSPolicy_Resource_FileResource_PRESENT,
										Path:  wantPaths[1],
										Source: &osconfigpb.OSPolicy_Resource_FileResource_File{
											File: &osconfigpb.OSPolicy_Resource_File{
												Type: &osconfigpb.OSPolicy_Resource_File_Gcs_{
													Gcs: &osconfigpb.OSPolicy_Resource_File_Gcs{
														Bucket:     testResourceBucket,
														Object:     "OSPolicies/test_file",
														Generation: 1617666133905437,
													},
												},
											},
										},
									},
								},
							},
							{
								Id: "file-present-from-uri",
								ResourceType: &osconfigpb.OSPolicy_Resource_File_{
									File: &osconfigpb.OSPolicy_Resource_FileResource{
										State: osconfigpb.OSPolicy_Resource_FileResource_PRESENT,
										Path:  wantPaths[2],
										Source: &osconfigpb.OSPolicy_Resource_FileResource_File{
											File: &osconfigpb.OSPolicy_Resource_File{
												Type: &osconfigpb.OSPolicy_Resource_File_Remote_{
													Remote: &osconfigpb.OSPolicy_Resource_File_Remote{
														Uri:            "https://storage.googleapis.com/osconfig-agent-end2end-test-resources/OSPolicies/test_file",
														Sha256Checksum: "e0ef7229e64c61596d8be928397e19fcc542ac920c4132106fb1ec2295dd73d1",
													},
												},
											},
										},
									},
								},
							},
							{
								Id: "file-present-from-local",
								ResourceType: &osconfigpb.OSPolicy_Resource_File_{
									File: &osconfigpb.OSPolicy_Resource_FileResource{
										State: osconfigpb.OSPolicy_Resource_FileResource_PRESENT,
										Path:  wantPaths[3],
										Source: &osconfigpb.OSPolicy_Resource_FileResource_File{
											File: &osconfigpb.OSPolicy_Resource_File{
												Type: &osconfigpb.OSPolicy_Resource_File_LocalPath{
													LocalPath: wantPaths[0],
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId: "files-absent",
			State:      osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: []*osconfigpb.OSPolicyResourceCompliance{
				{
					OsPolicyResourceId: "file-absent",
					ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
					},
					State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
			},
		},
		{
			OsPolicyId: "files-present",
			State:      osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: []*osconfigpb.OSPolicyResourceCompliance{
				{
					OsPolicyResourceId: "file-present-from-content",
					ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
					},
					State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "file-present-from-gcs",
					ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
					},
					State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "file-present-from-uri",
					ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
					},
					State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "file-present-from-local",
					ConfigSteps: []*osconfigpb.OSPolicyResourceConfigStep{
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
						{
							Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
							Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
						},
					},
					State: osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
			},
		},
	}
	ss := getStartupScriptFile(name, pkgManager, dnePath, wantPaths)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{fileDNE}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func addFileResourceTests(key string) []*osPolicyTestSetup {
	var pkgTestSetup []*osPolicyTestSetup
	for name, image := range utils.HeadAptImages {
		pkgTestSetup = append(pkgTestSetup, buildFileResourceTests(name, image, "apt", key))
	}
	for name, image := range utils.HeadELImages {
		pkgTestSetup = append(pkgTestSetup, buildFileResourceTests(name, image, "yum", key))
	}
	for name, image := range utils.HeadSUSEImages {
		pkgTestSetup = append(pkgTestSetup, buildFileResourceTests(name, image, "zypper", key))
	}
	for name, image := range utils.HeadWindowsImages {
		pkgTestSetup = append(pkgTestSetup, buildFileResourceTests(name, image, "googet", key))
	}
	return pkgTestSetup
}

func buildLinuxExecResourceTests(name, image, pkgManager, key string) *osPolicyTestSetup {
	assertTimeout := 180 * time.Second
	testName := linuxExecResource
	machineType := "e2-medium"
	checkPaths := []string{"/path1", "/path2", "/path3", "/path4"}

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				// Each of these resources checks for and then creates a file from checkPaths.
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "exec-gcs",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Gcs_{
														Gcs: &osconfigpb.OSPolicy_Resource_File_Gcs{
															Bucket:     "osconfig-agent-end2end-test-resources",
															Object:     "OSPolicies/validate_shell",
															Generation: 1618251880260222,
														},
													},
												},
											},
											Args:        []string{checkPaths[0]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Gcs_{
														Gcs: &osconfigpb.OSPolicy_Resource_File_Gcs{
															Bucket:     "osconfig-agent-end2end-test-resources",
															Object:     "OSPolicies/enforce_none",
															Generation: 1618246663082861,
														},
													},
												},
											},
											Args:        []string{checkPaths[0]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
							{
								Id: "exec-uri",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Remote_{
														Remote: &osconfigpb.OSPolicy_Resource_File_Remote{
															Uri:            "https://storage.googleapis.com/osconfig-agent-end2end-test-resources/OSPolicies/validate_shell",
															Sha256Checksum: "67837bf4b3be1ff84758e22d2eb46db4904dd57eb50f19de3f51a52be1c5b555",
														},
													},
												},
											},
											Args:        []string{checkPaths[1]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Remote_{
														Remote: &osconfigpb.OSPolicy_Resource_File_Remote{
															Uri:            "https://storage.googleapis.com/osconfig-agent-end2end-test-resources/OSPolicies/enforce_none",
															Sha256Checksum: "6b9f3936ddc557819281d9bd933aa42be513d31447ca683a9bc26e9f14d6abf1",
														},
													},
												},
											},
											Args:        []string{checkPaths[1]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
							// These local scripts are created by the startup script.
							{
								Id: "exec-local",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_LocalPath{
														LocalPath: "/validate_shell",
													},
												},
											},
											Args:        []string{checkPaths[2]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_LocalPath{
														LocalPath: "/enforce_none",
													},
												},
											},
											Args:        []string{checkPaths[2]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
							{
								Id: "exec-script",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_Script{
												Script: "if ls $1 >/dev/null; then\nexit 100\nfi\nexit 101",
											},
											Args:        []string{checkPaths[3]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_Script{
												Script: "#!/bin/sh\ntouch $1\nexit 100",
											},
											Args:        []string{checkPaths[3]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	expectedSteps := []*osconfigpb.OSPolicyResourceConfigStep{
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId: testName,
			State:      osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: []*osconfigpb.OSPolicyResourceCompliance{
				{
					OsPolicyResourceId: "exec-gcs",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-uri",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-local",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-script",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
			},
		},
	}
	ss := getStartupScriptExec(name, pkgManager, checkPaths)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{fileExists}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func buildWindowsExecResourceTests(name, image, pkgManager, key string) *osPolicyTestSetup {
	assertTimeout := 180 * time.Second
	testName := windowsExecResource
	machineType := "e2-standard-2"
	checkPaths := []string{"/path1", "/path2", "/path3", "/path4", "/path5", "/path6"}

	instanceName := fmt.Sprintf("%s-%s-%s-%s", path.Base(name), testName, key, utils.RandString(3))
	ospa := &osconfigpb.OSPolicyAssignment{
		InstanceFilter: &osconfigpb.OSPolicyAssignment_InstanceFilter{
			InclusionLabels: []*osconfigpb.OSPolicyAssignment_LabelSet{{
				Labels: map[string]string{"name": instanceName}},
			},
		},
		Rollout: &osconfigpb.OSPolicyAssignment_Rollout{
			DisruptionBudget: &osconfigpb.FixedOrPercent{Mode: &osconfigpb.FixedOrPercent_Percent{Percent: 100}},
			MinWaitDuration:  &durationpb.Duration{Seconds: 0},
		},
		OsPolicies: []*osconfigpb.OSPolicy{
			{
				Id:   testName,
				Mode: osconfigpb.OSPolicy_ENFORCEMENT,
				// Each of these resources checks for and then creates a file from checkPaths.
				ResourceGroups: []*osconfigpb.OSPolicy_ResourceGroup{
					{
						Resources: []*osconfigpb.OSPolicy_Resource{
							{
								Id: "exec-gcs-cmd",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Gcs_{
														Gcs: &osconfigpb.OSPolicy_Resource_File_Gcs{
															Bucket:     "osconfig-agent-end2end-test-resources",
															Object:     "OSPolicies/validate.cmd",
															Generation: 1618340275278922,
														},
													},
												},
											},
											Args:        []string{checkPaths[0]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Gcs_{
														Gcs: &osconfigpb.OSPolicy_Resource_File_Gcs{
															Bucket:     "osconfig-agent-end2end-test-resources",
															Object:     "OSPolicies/enforce.cmd",
															Generation: 1618340277107450,
														},
													},
												},
											},
											Args:        []string{checkPaths[0]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
							{
								Id: "exec-uri-cmd",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Remote_{
														Remote: &osconfigpb.OSPolicy_Resource_File_Remote{
															Uri:            "https://storage.googleapis.com/osconfig-agent-end2end-test-resources/OSPolicies/validate.cmd",
															Sha256Checksum: "1635e97b142fa9dd21bb023093ede409d242f52a535ad779bb80539db95c8f77",
														},
													},
												},
											},
											Args:        []string{checkPaths[1]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Remote_{
														Remote: &osconfigpb.OSPolicy_Resource_File_Remote{
															Uri:            "https://storage.googleapis.com/osconfig-agent-end2end-test-resources/OSPolicies/enforce.cmd",
															Sha256Checksum: "d9eef617ac54a8aa46be87b652a5e70b1b32035d1540e8d9bdae5dff406ceca5",
														},
													},
												},
											},
											Args:        []string{checkPaths[1]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
							// These local scripts are created by the startup script.
							{
								Id: "exec-local-cmd",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_LocalPath{
														LocalPath: "/validate.cmd",
													},
												},
											},
											Args:        []string{checkPaths[2]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_LocalPath{
														LocalPath: "/enforce.cmd",
													},
												},
											},
											Args:        []string{checkPaths[2]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_NONE,
										},
									},
								},
							},
							// No support for executing a script with no Interpreter via Script on Windows.
							{
								Id: "exec-script",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_Script{
												Script: "if exists %1 exit 100\nexit 101",
											},
											Args:        []string{checkPaths[3]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_SHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_Script{
												Script: "New-Item -ItemType File -Path $Args[0]; exit 100",
											},
											Args:        []string{checkPaths[3]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_POWERSHELL,
										},
									},
								},
							},
							// Windows does not support executing a powershell script directly.
							{
								Id: "exec-gcs-uri-ps1",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Gcs_{
														Gcs: &osconfigpb.OSPolicy_Resource_File_Gcs{
															Bucket:     "osconfig-agent-end2end-test-resources",
															Object:     "OSPolicies/validate.ps1",
															Generation: 1617995966532645,
														},
													},
												},
											},
											Args:        []string{checkPaths[4]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_POWERSHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_Remote_{
														Remote: &osconfigpb.OSPolicy_Resource_File_Remote{
															Uri:            "https://storage.googleapis.com/osconfig-agent-end2end-test-resources/OSPolicies/enforce.ps1",
															Sha256Checksum: "a5737e35f8a3a04785e4e0b9ffa90c5209db320c0ef9692672f5fb0b1dfe99d2",
														},
													},
												},
											},
											Args:        []string{checkPaths[4]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_POWERSHELL,
										},
									},
								},
							},
							{
								Id: "exec-local-powershell",
								ResourceType: &osconfigpb.OSPolicy_Resource_Exec{
									Exec: &osconfigpb.OSPolicy_Resource_ExecResource{
										Validate: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_LocalPath{
														LocalPath: "/validate.ps1",
													},
												},
											},
											Args:        []string{checkPaths[5]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_POWERSHELL,
										},
										Enforce: &osconfigpb.OSPolicy_Resource_ExecResource_Exec{
											Source: &osconfigpb.OSPolicy_Resource_ExecResource_Exec_File{
												File: &osconfigpb.OSPolicy_Resource_File{
													Type: &osconfig.OSPolicy_Resource_File_LocalPath{
														LocalPath: "/enforce.ps1",
													},
												},
											},
											Args:        []string{checkPaths[5]},
											Interpreter: osconfigpb.OSPolicy_Resource_ExecResource_Exec_POWERSHELL,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	expectedSteps := []*osconfigpb.OSPolicyResourceConfigStep{
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_VALIDATION,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_ENFORCEMENT,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
		{
			Type:    osconfigpb.OSPolicyResourceConfigStep_DESIRED_STATE_CHECK_POST_ENFORCEMENT,
			Outcome: osconfigpb.OSPolicyResourceConfigStep_SUCCEEDED,
		},
	}
	wantCompliances := []*osconfigpb.InstanceOSPoliciesCompliance_OSPolicyCompliance{
		{
			OsPolicyId: testName,
			State:      osconfigpb.OSPolicyComplianceState_COMPLIANT,
			OsPolicyResourceCompliances: []*osconfigpb.OSPolicyResourceCompliance{
				{
					OsPolicyResourceId: "exec-gcs-cmd",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-uri-cmd",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-local-cmd",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-script",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-gcs-uri-ps1",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
				{
					OsPolicyResourceId: "exec-local-powershell",
					ConfigSteps:        expectedSteps,
					State:              osconfigpb.OSPolicyComplianceState_COMPLIANT,
				},
			},
		},
	}
	ss := getStartupScriptExec(name, pkgManager, checkPaths)
	return newOsPolicyTestSetup(image, name, instanceName, testName, []string{fileExists}, machineType, ospa, ss, assertTimeout, wantCompliances)
}

func addExecResourceTests(key string) []*osPolicyTestSetup {
	var pkgTestSetup []*osPolicyTestSetup
	for name, image := range utils.HeadAptImages {
		pkgTestSetup = append(pkgTestSetup, buildLinuxExecResourceTests(name, image, "apt", key))
	}
	for name, image := range utils.HeadELImages {
		pkgTestSetup = append(pkgTestSetup, buildLinuxExecResourceTests(name, image, "yum", key))
	}
	for name, image := range utils.HeadSUSEImages {
		pkgTestSetup = append(pkgTestSetup, buildLinuxExecResourceTests(name, image, "zypper", key))
	}
	for name, image := range utils.HeadWindowsImages {
		pkgTestSetup = append(pkgTestSetup, buildWindowsExecResourceTests(name, image, "googet", key))
	}
	return pkgTestSetup
}

func generateAllTestSetup() []*osPolicyTestSetup {
	key := utils.RandString(3)

	pkgTestSetup := []*osPolicyTestSetup{}
	pkgTestSetup = append(pkgTestSetup, addPackageResourceTests(key)...)
	pkgTestSetup = append(pkgTestSetup, addRepositoryResourceTests(key)...)
	pkgTestSetup = append(pkgTestSetup, addFileResourceTests(key)...)
	pkgTestSetup = append(pkgTestSetup, addExecResourceTests(key)...)
	return pkgTestSetup
}