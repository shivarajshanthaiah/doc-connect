package routes

import (
	"doc-connect/authentication"
	"doc-connect/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes() *gin.Engine {
	//creates a new Gin engine instance with default configurations
	r := gin.Default()

	//user routers
	r.POST("/users/login", controllers.PatientLogin)
	r.POST("/users/signup", controllers.PatientSignup)
	r.POST("/users/verify", controllers.UserOtpVerify)
	r.GET("/pay/invoice/online", controllers.MakePaymentOnline)
	r.GET("/payment/success", controllers.SuccessPage)

	user := r.Group("/user")
	user.Use(authentication.PatientAuthMiddleware())
	{
		user.GET("/doctors/:doctor_id/available-slots", controllers.GetAvailableTimeSlots)
		user.GET("/logout", controllers.PatientLogout)
		user.GET("/doctor/:specialization", controllers.GetDoctorsBySpeciality)
		user.POST("/book/appointment", controllers.BookAppointment)
		user.POST("/pay/invoice/offline", controllers.PayInvoiceOffline)
		user.GET("/wallet/:userid", controllers.Wallet)
		user.POST("/cancel/appointment/:id", controllers.CancelAppointment)
		user.GET("/appointment/history/:id", controllers.GetAppointmenentHistory)
		user.POST("/pay/invoice/wallet", controllers.PayFromWallet)

	}

	//Admin routes

	r.POST("/admin/login", controllers.AdminLogin)

	admin := r.Group("/admin")
	admin.Use(authentication.AdminAuthMiddleware())
	{
		admin.POST("/logout", controllers.AdminLogout)
		admin.GET("/view/hospitals", controllers.ViewHospitals)
		admin.POST("/add/hospital", controllers.AddHospital)
		admin.GET("/search/hospital/:id", controllers.SearchHospital)
		admin.PATCH("/update/hospital/:id", controllers.UpdateHospital)
		admin.POST("/remove/hospital/:id", controllers.RemoveHospital)
		admin.GET("/view/deleted/hospitals", controllers.ViewDeletedHospitals)
		admin.GET("/view/Active/hospitals", controllers.ViewActiveHospitals)
		admin.POST("/verify/doctor/:id", controllers.UpdateDoctor)
		admin.GET("/view/verified/doctors", controllers.ViewVerifiedDoctors)
		admin.GET("/view/doctor/:id", controllers.GetDoctorByID)
		admin.GET("/view/doctors/:specialization", controllers.GetDoctorBySpeciality)
		admin.GET("/view/notVerified/doctors", controllers.ViewNotVerifiedDoctors)
		admin.GET("/view/verified/approved/doctors", controllers.ViewVerifiedApprovedDoctors)
		admin.GET("/view/verified/notApproved/doctors", controllers.ViewVerifiedNotApprovedDoctors)
		admin.GET("/view/invoice", controllers.GetInvoice)
		admin.GET("/total/appointments", controllers.GetBookingStatusCounts)
		admin.GET("/doctor-wise/bookings", controllers.GetDoctorWiseBookings)
		admin.GET("/department-wise/bookings", controllers.GetDepartmentWiseBookings)
		admin.GET("/total/revenue", controllers.GetTotalRevenue)
		admin.GET("/revenue/startdate", controllers.GetSpecificRevenue)
	}

	//Doctor routes
	r.POST("doctor/signup", controllers.Signup)
	r.POST("doctor/verify", controllers.VerifyOTP)
	r.GET("view/hospitals", controllers.ViewHospital)
	//r.POST("doctor/signup", doctorControllers.DoctorSignup)
	r.POST("/doctor/login", controllers.DoctorLogin)

	doctors := r.Group("/doctor")
	doctors.Use(authentication.DoctorAuthMiddleware())
	{
		doctors.POST("/update/availability", controllers.SaveAvailability)
		doctors.GET("/logout", controllers.DoctorLogout)
		doctors.POST("/add/prescription", controllers.AddPrescription)
		doctors.POST("/cancel/appointment/:id", controllers.CancelAppointment)
		doctors.GET("/appointment/history/:id", controllers.GetAppHistory)
		doctors.GET("/appointment/:doctor_id/date", controllers.GetDoctorAppointmentsByDate)
	}

	return r
}
