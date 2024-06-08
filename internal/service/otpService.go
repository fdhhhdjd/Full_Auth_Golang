package service

import (
	"time"

	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/configs/common/constants"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/global"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/internal/models"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/internal/repo"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/pkg/helpers"
	"github.com/fdhhhdjd/Go_Secure_Auth_Pro/response"
	"github.com/gin-gonic/gin"
)

// SendOtp generates and sends an OTP (One-Time Password) to the user.
// It retrieves the user information from the request context, generates an OTP,
// saves it in the database, and returns a response containing the OTP details.
func SendOtp(c *gin.Context, userId int, time time.Time) *models.SendOtpResponse {
	otp := helpers.GenerateOTP(6)
	timeExpired := time
	resultOtp, err := repo.CreateOtp(global.DB, models.CreateOtpParams{
		UserID:    userId,
		OtpCode:   otp,
		ExpiresAt: timeExpired,
	})

	if err != nil {
		return nil
	}

	return &models.SendOtpResponse{
		Id:        resultOtp.UserID,
		Code:      resultOtp.OtpCode,
		ExpiredAt: timeExpired.String(),
	}
}

// VerificationOtp handles the verification of OTP (One-Time Password) for user login.
// It takes a gin.Context object as a parameter and returns a pointer to models.LoginResponse.
// The function first binds the JSON request body to the models.OtpRequest struct.
// If there is an error in binding, it returns a bad request error response.
// Then, it retrieves the new OTPs from the database using the repo.GetNewOtps function.
// If there is an error in retrieving the OTPs or if no OTP is found, it returns a bad request error response.
// Otherwise, it updates the OTP's IsActive field to false using the repo.UpdateOtpIsActive function.
// Next, it creates an access token, refetch token, and encodes the public key using the createKeyAndToken function.
// If any of these values are empty, it returns a bad request error response.
// It then updates the user's device information using the upsetDevice function.
// Finally, it sets a cookie with the refetch token and returns a LoginResponse object with the user's ID, device ID, email, and access token.
//
// @Summary Verify OTP
// @Description Handles the OTP verification process
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.OtpRequest true "OTP verification request body"
// @Param X-Device-Id header string true "Device ID"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /auth/verify-otp [post]
func VerificationOtp(c *gin.Context) *models.LoginResponse {
	var req models.OtpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequestError(c)
		return nil
	}

	otp, err := repo.GetNewOtps(global.DB, req.Otp)

	if err != nil {
		response.BadRequestError(c)
		return nil
	}

	if len(otp) == 0 {
		response.BadRequestError(c)
		return nil
	}

	resultInfo := &otp[0]

	repo.UpdateOtpIsActive(global.DB, models.UpdateOtpIsActiveParams{IsActive: false, OtpCode: otp[0].OtpCode})

	accessToken, refetchToken, resultEncodePublicKey := createKeyAndToken(models.UserIDEmail{
		ID:    resultInfo.ID,
		Email: resultInfo.Email,
	})

	if accessToken == "" || refetchToken == "" || resultEncodePublicKey == "" {
		response.BadRequestError(c)
		return nil
	}

	resultInfoDevice := upsetDevice(c, resultInfo.ID, resultEncodePublicKey)

	setCookie(c, constants.UserLoginKey, refetchToken, "/", constants.AgeCookie)

	return &models.LoginResponse{
		ID:          resultInfo.ID,
		DeviceID:    resultInfoDevice.DeviceID,
		Email:       resultInfo.Email,
		AccessToken: accessToken,
	}
}
