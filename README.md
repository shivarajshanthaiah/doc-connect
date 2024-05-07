# Go doconnect

This project is of Doctor Management System where users/patients can do the booking of thier consultaion with respective doctor. The backend project is implemented in Go language using the Gin web framework, PostgreSQL for data storage, Redis for caching, Twilio for SMS OTP verificaton, Razorpay for payment processing, SMTP for email notifications, Gomail for email notification with PDF attachment and GORM as the ORM.

## Live Demo

Access the live demo of this project at (https://godoconnect.life).

## Features


- Added SMS OTP verification for user.
- Added E-Mail OTP verification for doctors.
- Added wallet feature.
  - Can pay from the wallet if balance is sufficient
- Proper appoinmtens conflict handling:
  - No dobuble bookings.
  - No duplicate bookings.
  - Confirmation only after payment.
- Invoice generation after succesfull appointment booking.
- Inoive is sent through email with PDF attachment.
- Admin routes for overall controlls.
- Doctor routes for adding prescription, updating avilability, etc.


### User Features

- **Authentication:**
  - User registration with SMS OTP verification for enhanced security.
  - Login with credentials.

- **Appointment Management:**
  - View available time slots.
  - Search and view doctors by speciality.
  - Proper error handling.

### Admin Features

- **User and Role Management:**
  - Manage Users.
  - Manage Doctors, Hospitals.
  - Dashboard with access to every ongoing information.
  - Verifying dcotors and hospitals.

### Doctor Features

- **Appointment Management:**
  - View appointment details - History and Sheduled.
  - Update prescription.


## Technologies Used

- **Backend:**
  - Go language
  - Gin web framework
  - PostgreSQL for data storage
  - Redis for caching
  - Twilio for SMS OTP verification
  - Razorpay for payment processing
  - SMTP for email notifications
  - Go-Mail for pdf attachment
  - GORM as the ORM

## API Documentation

For detailed API documentation, refer to - (https://documenter.getpostman.com/view/32823353/2sA3JJA3bX) - API Documentation.


## Setup and Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/your-username/library-management.git

2.Install dependencies:

    cd doc-connect
    go get -u github.com/gin-gonic/gin
    go get -u github.com/razorpay/razorpay-go
    go get -u github.com/go-redis/redis
    go get -u gorm.io/driver/postgres
    go get -u github.com/twilio/twilio-go

3.Set up the database:

    CREATE DATABASE docapp

4.Configure environment variables:

    DB_Config="host=localhost user=##### password=***** dbname=docapp port=0000 sslmode=disable"  


    Email="youremail@email.com"
    Password="__ __ __ __(use app password)"
    
    
    RAZORPAY_KEY_ID="________________(apikey)"
    RAZORPAY_SECRET="_______________(api secret)"


    TWILIO_ACCOUNT_SID="___________________"
    TWILIO_AUTHTOKEN="_____________________"
    TWILIO_SERVIES_ID="___________________"
    TWILIO_PHONENUMBER="__________________(twilio phone number)"

5.Run the application:

    make run
