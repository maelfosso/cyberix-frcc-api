package jobs

func (r *Runner) registerJobs() {
	SendVerificationEmail(r, r.emailer)
	SendOtpEmail(r, r.emailer)
	SendWelcomeEmail(r, r.emailer)
}
