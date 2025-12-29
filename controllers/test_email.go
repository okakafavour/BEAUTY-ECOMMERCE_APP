package controllers

import (
	"net/http"

	"beauty-ecommerce-backend/utils"

	"github.com/gin-gonic/gin"
)

func SendProofEmail(c *gin.Context) {
	utils.QueueEmail(
		"okakafavour81@gmail.com", // YOUR gmail
		"Favour",
		"Service Subscription Expiration Notice",
		`
		<p>Hello,</p>
		<p>This email is to formally document that the following third-party services used in the Beauty E-commerce application are <b>paid services</b> and are approaching expiration:</p>
		<ul>
			<li><b>Cloudinary</b> – image uploads & storage</li>
			<li><b>Database hosting</b> – users, orders & system data</li>
		</ul>
		<p>If these subscriptions are not renewed, the application may experience downtime or data issues.</p>
		<p>This message serves as official notice and proof.</p>
		<p>Regards,<br/> @Gateway.com</p>
		`,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Proof email queued",
	})
}
