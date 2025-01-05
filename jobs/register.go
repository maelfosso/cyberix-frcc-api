package jobs

func (r *Runner) registerJobs() {
	SendVerificationEmail(r, r.emailer)
}
