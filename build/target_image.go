package build

// TargetImage wraps a [Target] and provides builder style methods for altering its
// internal image configuration.
type TargetImage struct {
	target *Target
}

// Entrypoint sets the runtime entrypoint of the image.
func (img *TargetImage) Entrypoint(entrypoint []string) *TargetImage {
	img.target.image.Config.Entrypoint = entrypoint
	return img
}

// User sets the runtime username or UID of the image.
func (img *TargetImage) User(user string) *TargetImage {
	img.target.image.Config.User = img.target.ExpandEnv(user)
	return img
}

// WorkingDirectory sets the runtime working directory of the image.
func (img *TargetImage) WorkingDirectory(path string) *TargetImage {
	img.target.image.Config.WorkingDir = img.target.ExpandEnv(path)
	return img
}

// AddEnv appends the runtime environment variables of the image.
func (img *TargetImage) AddEnv(env map[string]string) *TargetImage {
	if img.target.image.Config.Env == nil {
		img.target.image.Config.Env = []string{}
	}

	for _, k := range sortedKeys(env) {
		img.target.image.Config.Env = append(
			img.target.image.Config.Env,
			k+"="+quote(img.target.ExpandEnv(env[k])),
		)
	}

	return img
}

// AddLabels appends the image configuration labels.
func (img *TargetImage) AddLabels(labels map[string]string) *TargetImage {
	if img.target.image.Config.Labels == nil {
		img.target.image.Config.Labels = map[string]string{}
	}

	for _, k := range sortedKeys(labels) {
		img.target.image.Config.Labels[k] = img.target.ExpandEnv(labels[k])
	}

	return img
}
