package utils

import (
	"fmt"

	"tayaria-warranty-be/models"

	"gopkg.in/gomail.v2"
)

// SendWarrantyConfirmationEmail sends a confirmation email to the user when a warranty is registered
func SendWarrantyConfirmationEmail(warranty models.Warranty) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "contact.tayaria@kitloongholdings.com")
	m.SetHeader("To", warranty.Email)
	m.SetHeader("Subject", "Warranty Registration Confirmation - Tayaria")

	// Create email body with warranty details and important information
	body := fmt.Sprintf(`
Dear %s,

🎉 Thank you for choosing Tayaria! Your warranty registration has been successfully completed.

📋 WARRANTY DETAILS:
• Car Plate: %s
• Purchase Date: %s
• Expiry Date: %s

⚠️ IMPORTANT WARRANTY TERMS:
1) Valid until 6 months from the date of purchase
2) Valid only if tyre has above 6mm of tread depth left
3) Valid only after a minimum purchase of 2 pcs in single receipt
4) Valid only for digital receipt
5) Invalid for tyre damages that are beyond repair

🔧 Need to file a claim? Head down to your nearest Tayaria shop:
https://tayaria.com/where-to-buy/?search=Kuala+Lumpur%%2CFederal+Territory+of+Kuala+Lumpur%%2CMalaysia

💡 Learn more about us: https://tayaria.com/

🚗 Explore our premium tyre collection: https://tayaria.com/brands/

If you have any questions, please don't hesitate to contact us at contact.tayaria@kitloongholdings.com

Warm regards,
The Tayaria Team 🛞
`, warranty.Name, warranty.CarPlate,
		warranty.PurchaseDate.Format("January 2, 2006"),
		warranty.ExpiryDate.Format("January 2, 2006"))

	m.SetBody("text/plain", body)

	// Configure SMTP dialer
	d := gomail.NewDialer("mail.kitloongholdings.com", 587, "contact.tayaria@kitloongholdings.com", "#Temp0000")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send warranty confirmation email: %w", err)
	}

	return nil
}
