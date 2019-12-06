package main

import (
	"fmt"
	"log"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// Analyze the Pod Spec and add to the record
func collectPodSpec(ps v1.PodSpec, r *Record) {

	log.Println("Analyzing the Pod Spec")

	// ATTRIBUTE - Handle HOST NETWORK ACCESS
	pHostNetwork := NewAttribute("Host Network", ps.HostNetwork, "")
	if pHostNetwork.Value == true {
		pHostNetwork.Flagged = true
		pHostNetwork.Reason = "Host (node) network access is set to [TRUE]"
	}
	r.Attributes = append(r.Attributes, *pHostNetwork)

	// ATTRIBUTE - Handle DNS Policy
	pDNS := NewAttribute("DNS Policy", ps.DNSPolicy, "")
	r.Attributes = append(r.Attributes, *pDNS)

	// ATTRIBUTE - Handle Workload level RUN AS USER
	pRunAsUser := NewAttribute("Run As User", ps.SecurityContext.RunAsUser, "")

	if pRunAsUser.Value == 0 { // checking if the defined user is root (move to analysis stage?)
		pRunAsUser.Flagged = true
		pRunAsUser.Reason = "Run as user is set to ROOT"
	}
	r.Attributes = append(r.Attributes, *pRunAsUser)

	// ATTRIBUTE - Handle Workload level RUN AS NON ROOT
	pNonRoot := NewAttribute("Run As Non Root", ps.SecurityContext.RunAsNonRoot, "")

	if pNonRoot.Value == false {
		pNonRoot.Flagged = true
		pNonRoot.Reason = "Run as non root is set to FALSE"
	}
	r.Attributes = append(r.Attributes, *pNonRoot)

	// ATTRIBUTE - Handle Workload level RUN AS GROUP
	pRunAsGroup := NewAttribute("Run As Group", ps.SecurityContext.RunAsGroup, "")

	if pRunAsGroup.Value == 0 {
		pRunAsGroup.Flagged = true
		pRunAsGroup.Reason = "Run as group is set to ROOT group"
	}
	r.Attributes = append(r.Attributes, *pRunAsGroup)

	// Keep the deployment level attributes for later use
	if pRunAsUser.Set {
		r.Global["Run As User"] = pRunAsUser.Value
	}

	if pRunAsGroup.Set {
		r.Global["Run As Group"] = pRunAsGroup.Value
	}

	if pNonRoot.Set {
		r.Global["Run As Non Root"] = pNonRoot.Value
	}

	// ATTRIBUTE - Handle POD Service Account Name
	pSA := NewAttribute("POD ServiceAccount", ps.ServiceAccountName, "")

	if pSA.Value == "default" {
		pSA.Flagged = true
		pSA.Reason = "POD default service account is set to [default]"
	}
	r.Attributes = append(r.Attributes, *pSA)

	log.Println("Done analyzing the Pod Spec")
}

func collectContainerPodSpec(c v1.Container, r *Record) {
	log.Printf("Working on container [%s]\n", c.Name)

	// ATTRIBUTE - Handle Image pull policy
	cPullPolicy := NewAttribute("Image Pull Policy", c.ImagePullPolicy, c.Name)

	r.Attributes = append(r.Attributes, *cPullPolicy)

	// ATTRIBUTE - Hanfle Image name check
	cImage := NewAttribute("Image Name", c.Image, c.Name)

	if strings.HasSuffix(strings.ToLower(cImage.Value.(string)), ":latest") {
		cImage.Flagged = true
		cImage.Reason = "Image with [LATEST] tag is not recommended"
	}
	r.Attributes = append(r.Attributes, *cImage)

	if c.SecurityContext != nil {
		// ATTRIBUTE - Handle Container privileged attribute
		cPriv := NewAttribute("Privileged", c.SecurityContext.Privileged, c.Name)

		if cPriv.Value == true {
			cPriv.Flagged = true
			cPriv.Reason = "Container set to allow [PRIVILEGED]"
		}
		r.Attributes = append(r.Attributes, *cPriv)

		// ATTRIBUTE - Handle Container AllowPrivilegeEscalation attribute
		cAllowPrivEsc := NewAttribute("Allow Privilege Escalation", c.SecurityContext.AllowPrivilegeEscalation, c.Name)

		if cAllowPrivEsc.Value == true {
			cAllowPrivEsc.Flagged = true
			cAllowPrivEsc.Default = false
			cAllowPrivEsc.Reason = "Container set to allow [PRIVILEGED]"
		}
		r.Attributes = append(r.Attributes, *cAllowPrivEsc)

		// ATTRIBUTE - Handle Read only root file system
		cROFS := NewAttribute("Read Only File System", c.SecurityContext.ReadOnlyRootFilesystem, c.Name)

		if cROFS.Value == false {
			cROFS.Flagged = true
			cROFS.Reason = "Container access to root file system is [NOT READ ONLY]"
		}
		r.Attributes = append(r.Attributes, *cROFS)

		// ATTRIBUTE - Handle Container Run As User
		cRunAsUser := NewAttribute("Run As User [CONTAINER]", c.SecurityContext.RunAsUser, c.Name)
		if cRunAsUser.Value == 0 {
			cRunAsUser.Flagged = true
			cRunAsUser.Reason = "Container Run As User is set to [ROOT]"
		}
		r.Attributes = append(r.Attributes, *cRunAsUser)

		// ATTRIBUTE - Handle Container Run As Group
		cRunAsGroup := NewAttribute("Run As Group [CONTAINER]", c.SecurityContext.RunAsGroup, c.Name)

		if cRunAsGroup.Value == 0 {
			cRunAsGroup.Flagged = true
			cRunAsGroup.Reason = "Container Run As Group is set to [ROOT]"
		}
		r.Attributes = append(r.Attributes, *cRunAsGroup)

	} else {
		log.Println("No security context for the container. Need to set defaults according to the workload or just defaults.")
	}

	fmt.Println("Done analyzing the Container in Pod Spec")
}
