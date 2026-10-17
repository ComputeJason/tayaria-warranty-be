package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"tayaria-warranty-be/db"
	"tayaria-warranty-be/models"

	"gopkg.in/gomail.v2"
)

// newMailDialer builds the SMTP dialer (Amazon SES) from SMTP_HOST, SMTP_PORT, SMTP_USERNAME
// and SMTP_PASSWORD. Port 587 automatically uses STARTTLS in gomail.v2
func newMailDialer() (*gomail.Dialer, error) {
	host := os.Getenv("SMTP_HOST")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	if host == "" || username == "" || password == "" {
		return nil, fmt.Errorf("SMTP_HOST, SMTP_USERNAME and SMTP_PASSWORD must be set")
	}
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		port = 587
	}
	return gomail.NewDialer(host, port, username, password), nil
}

// SendWarrantyConfirmationEmail sends a confirmation email to the user when a warranty is registered
func SendWarrantyConfirmationEmail(warranty models.Warranty) error {
	fmt.Printf("🔍 Starting to send warranty confirmation email to: %s\n", warranty.Email)

	m := gomail.NewMessage()
	m.SetHeader("From", "contact.tayaria@kitloongholdings.com")
	m.SetHeader("To", warranty.Email)
	m.SetHeader("Subject", "Warranty Registration Confirmation - Tayaria")

	// Create email body with warranty details and important information. Kept deliberately
	// plain (no emoji, minimal links, formal tone) and provided as both plain-text and HTML -
	// outgoing mail scanners (e.g. the SpamAssassin outgoing scan on mail.kitloongholdings.com)
	// score emoji-heavy/multi-link/HTML-only transactional mail harshly; see the 550 "classified
	// as SPAM" rejections this used to trigger.
	plainBody := fmt.Sprintf(`Dear %s,

Thank you for registering your tyre warranty with Tayaria. Your registration has been completed successfully.

Warranty Details:
Car Plate: %s
Purchase Date: %s
Expiry Date: %s

Warranty Terms:
1. Valid for 6 months from the date of purchase.
2. Valid only if the tyre has more than 6mm of tread depth remaining.
3. Valid only for a minimum purchase of 2 tyres on a single receipt.
4. Valid only with a digital receipt on file.
5. Valid only for tyre damage that is beyond repair.

If you need to file a claim, please visit your nearest Tayaria shop: https://tayaria.com/where-to-buy/

For any questions, please contact us at contact.tayaria@kitloongholdings.com.

Regards,
The Tayaria Team
`, warranty.Name, warranty.CarPlate,
		warranty.PurchaseDate.Format("January 2, 2006"),
		warranty.ExpiryDate.Format("January 2, 2006"))

	htmlBody := fmt.Sprintf(`
<html>
<body>
<p>Dear %s,</p>

<p>Thank you for registering your tyre warranty with Tayaria. Your registration has been completed successfully.</p>

<p><strong>Warranty Details:</strong><br>
Car Plate: %s<br>
Purchase Date: %s<br>
Expiry Date: %s</p>

<p><strong>Warranty Terms:</strong><br>
1. Valid for 6 months from the date of purchase.<br>
2. Valid only if the tyre has more than 6mm of tread depth remaining.<br>
3. Valid only for a minimum purchase of 2 tyres on a single receipt.<br>
4. Valid only with a digital receipt on file.<br>
5. Valid only for tyre damage that is beyond repair.</p>

<p>If you need to file a claim, please visit your nearest Tayaria shop: <a href="https://tayaria.com/where-to-buy/">https://tayaria.com/where-to-buy/</a></p>

<p>For any questions, please contact us at contact.tayaria@kitloongholdings.com.</p>

<p>Regards,<br>
The Tayaria Team</p>
</body>
</html>
`, warranty.Name, warranty.CarPlate,
		warranty.PurchaseDate.Format("January 2, 2006"),
		warranty.ExpiryDate.Format("January 2, 2006"))

	// Plain-text first, HTML as the richer alternative - the standard multipart/alternative
	// order, and itself a commonly-cited factor in outgoing spam scoring when missing entirely.
	m.SetBody("text/plain", plainBody)
	m.AddAlternative("text/html", htmlBody)

	fmt.Printf("📧 Email body prepared, attempting to send...\n")

	d, err := newMailDialer()
	if err != nil {
		return fmt.Errorf("failed to configure email: %w", err)
	}

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("❌ Failed to send email: %v\n", err)
		return fmt.Errorf("failed to send warranty confirmation email: %w", err)
	}

	fmt.Printf("✅ Email sent successfully to %s\n", warranty.Email)
	return nil
}

// SendClaimAcceptanceEmail sends a notification email to admin when a claim is accepted
func SendClaimAcceptanceEmail(claim *models.Claim) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "contact.tayaria@kitloongholdings.com")
	m.SetHeader("To", "warranty@kitloongholdings.com")
	m.SetHeader("Subject", "Claim Accepted - Tayaria Warranty")

	// Format claim date (date only, no time)
	claimDate := claim.CreatedAt.Format("January 2, 2006")

	// Get purchase date from warranty if available
	purchaseDate := claimDate // fallback to claim date
	if claim.WarrantyID != nil {
		warranty, err := db.GetWarrantyByID(*claim.WarrantyID)
		if err != nil {
			// Log error but continue with fallback date
			fmt.Printf("Failed to get warranty for purchase date: %v\n", err)
		} else {
			purchaseDate = warranty.PurchaseDate.Format("January 2, 2006")
		}
	}

	// Build tyre details section
	tyreDetailsSection := ""
	if len(claim.TyreDetails) > 0 {
		tyreDetailsSection = fmt.Sprintf("Quantity: %d\n\n", len(claim.TyreDetails))
		for i, tyre := range claim.TyreDetails {
			tyreDetailsSection += fmt.Sprintf("%d. Brand: %s\n   Size: %s\n   Tread Pattern: %s\n\n",
				i+1, tyre.Brand, tyre.Size, tyre.TreadPattern)
		}
	} else {
		tyreDetailsSection = "No tyre details available\n\n"
	}

	// Create email body
	body := fmt.Sprintf(`
<html>
<body>
  <h2>Claim Details</h2>
  <p><strong>Purchase date:</strong> %s</p>
  <p><strong>Claim date:</strong> %s</p>
  <p><strong>Shop name:</strong> %s</p>
  <p><strong>Shop contact:</strong> %s</p>
  <p><strong>Vehicle Carplate:</strong> %s</p>

  <h3>Tyre Claim Details</h3>
  <p><strong>Quantity:</strong> %d</p>
  <ul>
`, purchaseDate, claimDate, claim.ShopName, claim.Contact, claim.CarPlate, len(claim.TyreDetails))

	for _, tyre := range claim.TyreDetails {
		body += fmt.Sprintf(`
    <li>
      <strong>Brand:</strong> %s<br>
      <strong>Size:</strong> %s<br>
      <strong>Tread Pattern:</strong> %s
    </li>
  `, tyre.Brand, tyre.Size, tyre.TreadPattern)
	}

	body += `
  </ul>
</body>
</html>
`
	m.SetBody("text/html", body)

	d, err := newMailDialer()
	if err != nil {
		return fmt.Errorf("failed to configure email: %w", err)
	}

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send claim acceptance email: %w", err)
	}

	return nil
}

// SendClaimCreationEmail sends a notification email to admin when a new claim is created
func SendClaimCreationEmail(claim *models.Claim) error {
	fmt.Printf("Starting to send claim creation email for claim ID: %s\n", claim.ID)

	m := gomail.NewMessage()
	m.SetHeader("From", "contact.tayaria@kitloongholdings.com")
	m.SetHeader("To", "warranty@kitloongholdings.com")
	m.SetHeader("Subject", fmt.Sprintf("New Warranty Claim Submitted - Car Plate: %s - Action Required", claim.CarPlate))

	// Format claim date with time in Malaysia timezone (GMT+8)
	malaysiaLoc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	claimDate := claim.CreatedAt.In(malaysiaLoc).Format("January 2, 2006 at 3:04 PM")

	// Create email body
	body := fmt.Sprintf(`
<html>
<body>
  <h2>New Warranty Claim Submitted</h2>
  
  <p>A new warranty claim has been created and requires your attention.</p>
  
  <hr>
  
  <h3>Claim Information</h3>
  <p><strong>Status:</strong> Unacknowledged<br>
  <strong>Submitted:</strong> %s</p>
  
  <h3>Shop Details</h3>
  <p><strong>Shop Name:</strong> %s<br>
  <strong>Contact:</strong> %s</p>
  
  <h3>Customer Information</h3>
  <p><strong>Name:</strong> %s<br>
  <strong>Phone:</strong> %s<br>
  <strong>Email:</strong> %s<br>
  <strong>Vehicle Plate:</strong> %s</p>
  
  <hr>
  
  <h3>Next Steps</h3>
  <ol>
    <li>Review the claim details in the admin portal</li>
    <li>Verify warranty validity</li>
    <li>Change status to "Pending" to proceed with processing</li>
  </ol>
  
  <p>Best regards,<br>
  Tayaria Warranty System</p>
</body>
</html>
`, claimDate, claim.ShopName, claim.Contact, claim.CustomerName,
		claim.PhoneNumber, claim.Email, claim.CarPlate)

	m.SetBody("text/html", body)

	fmt.Printf("Email body prepared, attempting to send...\n")

	d, err := newMailDialer()
	if err != nil {
		return fmt.Errorf("failed to configure email: %w", err)
	}

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send claim creation email: %v\n", err)
		return fmt.Errorf("failed to send claim creation email: %w", err)
	}

	fmt.Printf("Claim creation email sent successfully for claim %s\n", claim.ID)
	return nil
}

// SendClaimRejectionEmail sends a notification email to admin when a claim is rejected
func SendClaimRejectionEmail(claim *models.Claim) error {
	fmt.Printf("Starting to send claim rejection email for claim ID: %s\n", claim.ID)

	m := gomail.NewMessage()
	m.SetHeader("From", "contact.tayaria@kitloongholdings.com")
	m.SetHeader("To", "warranty@kitloongholdings.com")
	m.SetHeader("Subject", fmt.Sprintf("Claim Rejected - Car Plate: %s", claim.CarPlate))

	// Format rejection date with time in Malaysia timezone (GMT+8)
	malaysiaLoc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	rejectionDate := claim.UpdatedAt.In(malaysiaLoc).Format("January 2, 2006 at 3:04 PM")

	// Create email body
	body := fmt.Sprintf(`
<html>
<body>
  <h2>Warranty Claim Rejected</h2>
  
  <p>A warranty claim has been rejected. This email serves as an internal record for your admin processes.</p>
  
  <hr>
  
  <h3>Claim Information</h3>
  <p><strong>Status:</strong> Rejected<br>
  <strong>Rejected On:</strong> %s</p>
  
  <h3>Shop Details</h3>
  <p><strong>Shop Name:</strong> %s<br>
  <strong>Contact:</strong> %s</p>
  
  <h3>Customer Information</h3>
  <p><strong>Name:</strong> %s<br>
  <strong>Phone:</strong> %s<br>
  <strong>Email:</strong> %s<br>
  <strong>Vehicle Plate:</strong> %s</p>
  
  <h3>Rejection Reason</h3>
  <p style="background-color: #f9f9f9; padding: 15px; border-left: 4px solid #e74c3c;">
  %s
  </p>
`, rejectionDate, claim.ShopName, claim.Contact, claim.CustomerName,
		claim.PhoneNumber, claim.Email, claim.CarPlate, claim.RejectionReason)

	// Add tyre details section if provided
	if len(claim.TyreDetails) > 0 {
		body += fmt.Sprintf(`
  <h3>Tyre Details</h3>
  <p><strong>Quantity:</strong> %d</p>
  <ul>
`, len(claim.TyreDetails))

		for _, tyre := range claim.TyreDetails {
			body += fmt.Sprintf(`
    <li>
      <strong>Brand:</strong> %s<br>
      <strong>Size:</strong> %s<br>
      <strong>Tread Pattern:</strong> %s
    </li>
`, tyre.Brand, tyre.Size, tyre.TreadPattern)
		}

		body += `
  </ul>
`
	}

	body += `
  <hr>
  
  <p><em>This is an automated notification for internal record-keeping.</em></p>
  
  <p>Best regards,<br>
  Tayaria Warranty System</p>
</body>
</html>
`

	m.SetBody("text/html", body)

	fmt.Printf("Email body prepared, attempting to send...\n")

	d, err := newMailDialer()
	if err != nil {
		return fmt.Errorf("failed to configure email: %w", err)
	}

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send claim rejection email: %v\n", err)
		return fmt.Errorf("failed to send claim rejection email: %w", err)
	}

	fmt.Printf("Claim rejection email sent successfully for claim %s\n", claim.ID)
	return nil
}
