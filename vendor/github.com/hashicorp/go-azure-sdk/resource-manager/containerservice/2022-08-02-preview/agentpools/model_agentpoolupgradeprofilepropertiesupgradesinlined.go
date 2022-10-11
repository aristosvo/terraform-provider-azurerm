package agentpools

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type AgentPoolUpgradeProfilePropertiesUpgradesInlined struct {
	IsPreview         *bool   `json:"isPreview,omitempty"`
	KubernetesVersion *string `json:"kubernetesVersion,omitempty"`
}
