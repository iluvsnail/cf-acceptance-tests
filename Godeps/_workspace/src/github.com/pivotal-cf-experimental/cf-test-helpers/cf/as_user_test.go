package cf_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/cf-test-helpers/cf"
	"github.com/vito/cmdtest"
)

var _ = Describe("AsUser", func() {
	var FakeThingsToRunAsUser = func() {}
	var FakeCfCalls = [][]string{}

	var FakeCf = func(args ...string) *cmdtest.Session {
		FakeCfCalls = append(FakeCfCalls, args)
		var session, _ = cmdtest.Start(exec.Command("echo", "nothing"))
		return session
	}
	var user = cf.NewUserContext("http://FAKE_API.example.com", "FAKE_USERNAME", "FAKE_PASSWORD", "FAKE_ORG", "FAKE_SPACE", true)

	BeforeEach(func() {
		FakeCfCalls = [][]string{}
		cf.Cf = FakeCf
	})

	It("calls cf api", func() {
		cf.AsUser(user, FakeThingsToRunAsUser)

		Expect(FakeCfCalls[0]).To(Equal([]string{"api", "http://FAKE_API.example.com", "--skip-ssl-validation"}))
	})

	It("calls cf auth", func() {
		cf.AsUser(user, FakeThingsToRunAsUser)

		Expect(FakeCfCalls[1]).To(Equal([]string{"auth", "FAKE_USERNAME", "FAKE_PASSWORD"}))
	})

	Describe("cf target", func() {
		Context("when org is set and space is set", func() {
			It("sets org and space", func() {
				cf.AsUser(user, FakeThingsToRunAsUser)

				Expect(FakeCfCalls[2]).To(Equal([]string{"target", "-o", "FAKE_ORG", "-s", "FAKE_SPACE"}))
			})

			Context("when org is set and space is NOT set", func() {
				BeforeEach(func() {
					user.Space = ""
				})

				It("sets org", func() {
					cf.AsUser(user, FakeThingsToRunAsUser)

					Expect(FakeCfCalls[2]).To(Equal([]string{"target", "-o", "FAKE_ORG"}))
				})

				Context("when org is NOT set and space is NOT set", func() {
					BeforeEach(func() {
						user.Org = ""
					})
				})
			})
		})

		Context("when org is NOT set", func() {

			It("sets org", func() {
				cf.AsUser(user, FakeThingsToRunAsUser)

				Expect(FakeCfCalls[2]).NotTo(Equal([]string{"target", "-o", "FAKE_ORG", "-s", "FAKE_SPACE"}))
			})
		})

		Context("when org is set", func() {
			BeforeEach(func() {
				user.Space = ""
			})

			Context("when space is set", func() {
			})
		})
	})

	It("calls cf logout", func() {
		cf.AsUser(user, FakeThingsToRunAsUser)

		Expect(FakeCfCalls[len(FakeCfCalls)-1]).To(Equal([]string{"logout"}))
	})

	It("logs out even if actions contain a failing expectation", func() {
		RegisterFailHandler(func(message string, callerSkip ...int) {})
		cf.AsUser(user, func() { Expect(1).To(Equal(2)) })
		RegisterFailHandler(Fail)

		Expect(FakeCfCalls[len(FakeCfCalls)-1]).To(Equal([]string{"logout"}))
	})

	It("calls the passed function", func() {
		called := false
		cf.AsUser(user, func() { called = true })

		Expect(called).To(BeTrue())
	})

	It("sets a unique CF_HOME value", func() {
		var (
			firstHome  string
			secondHome string
		)

		cf.AsUser(user, func() {
			firstHome = os.Getenv("CF_HOME")
		})

		cf.AsUser(user, func() {
			secondHome = os.Getenv("CF_HOME")
		})

		Expect(firstHome).NotTo(Equal(secondHome))
	})

	It("returns CF_HOME to its original value", func() {
		os.Setenv("CF_HOME", "some-crazy-value")
		cf.AsUser(user, func() {})

		Expect(os.Getenv("CF_HOME")).To(Equal("some-crazy-value"))
	})
})
