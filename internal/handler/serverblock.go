package handler

import (
	"fmt"
	webv1alpha1 "openresty-operator/api/v1alpha1"
	"openresty-operator/internal/utils"
	"strings"
)

func ValidateLocationRefs(locations map[string]*webv1alpha1.Location, locationRefs []string) (bool, []string) {
	var problems []string
	pathSeen := make(map[string]string)

	for _, refName := range locationRefs {
		loc := locations[refName]

		if loc == nil {
			problems = append(problems, fmt.Sprintf("Missing Location: %s", refName))
			continue
		}

		if !loc.Status.Ready {
			problems = append(problems, fmt.Sprintf("Location not ready: %s", refName))
		}

		for _, entry := range loc.Spec.Entries {
			path := entry.Path
			if otherRef, exists := pathSeen[path]; exists && otherRef != refName {
				problems = append(problems, fmt.Sprintf("Duplicated path '%s' in %s and %s", path, otherRef, refName))
			} else {
				pathSeen[path] = refName
			}
		}
	}

	return len(problems) == 0, problems
}

func GenerateServerBlockConfig(s *webv1alpha1.ServerBlock) string {
	var b strings.Builder

	b.WriteString("server {\n")
	b.WriteString(fmt.Sprintf("    listen %s;\n", s.Spec.Listen))

	serverName := fmt.Sprintf("%s.%s.svc.cluster.local", s.Name, s.Namespace)
	b.WriteString(fmt.Sprintf("    server_name %s;\n", serverName))

	for _, ref := range s.Spec.LocationRefs {
		includePath := fmt.Sprintf(utils.NginxLocationConfigDir+"/%s/%s.conf", ref, ref)
		b.WriteString(fmt.Sprintf("    include %s;\n", includePath))
	}

	for _, h := range s.Spec.Headers {
		b.WriteString(fmt.Sprintf("    add_header %s %s;\n", h.Key, h.Value))
	}

	for _, line := range s.Spec.Extra {
		b.WriteString("    " + line + "\n")
	}

	b.WriteString("}\n")
	return b.String()
}
