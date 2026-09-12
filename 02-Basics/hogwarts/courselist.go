//go:build !solution

package hogwarts

func GetCourseList(prereqs map[string][]string) []string {
	courses := make([]string, 0, len(prereqs))
	visited := make(map[string]struct{})
	visiting := make(map[string]struct{})

	var dfs func(string) bool

	dfs = func(course string) bool {
		if _, ok := visiting[course]; ok {
			return false
		}

		if _, ok := visited[course]; ok {
			return true
		}

		visiting[course] = struct{}{}

		for _, pre := range prereqs[course] {
			if !dfs(pre) {
				return false
			}
		}

		delete(visiting, course)
		visited[course] = struct{}{}
		courses = append(courses, course)

		return true
	}

	for course := range prereqs {
		if !dfs(course) {
			panic("cycle")
		}
	}

	return courses
}
