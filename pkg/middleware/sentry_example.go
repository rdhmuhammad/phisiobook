package middleware

import (
	"fmt"
	_ "github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// Example usage of Sentry error capturing in your controllers
func ExampleSentryUsage() {
	// This is just an example function showing how to use Sentry in your controllers

	// Example 1: Capturing an error in a Gin handler
	// func YourHandler(c *gin.Context) {
	//     // Your business logic here
	//     err := someBusinessLogic()
	//     if err != nil {
	//         // Capture error with enriched context
	//         CaptureError(c, err)
	//
	//         // Return error response
	//         c.JSON(500, gin.H{"error": "Internal server error"})
	//         return
	//     }
	//
	//     c.JSON(200, gin.H{"message": "Success"})
	// }

	// Example 2: Capturing a custom message
	// func AnotherHandler(c *gin.Context) {
	//     // Log important events
	//     CaptureMessage(c, "User performed important action", sentry.LevelInfo)
	//
	//     c.JSON(200, gin.H{"message": "Action completed"})
	// }

	// Example 3: Manual error capturing with additional context
	// func ComplexHandler(c *gin.Context) {
	//     hub := sentry.GetHubFromContext(c.Request.Context())
	//     if hub == nil {
	//         hub = sentry.CurrentHub()
	//     }
	//
	//     hub.WithScope(func(scope *sentry.Scope) {
	//         scope.SetTag("operation", "complex_business_logic")
	//         scope.SetExtra("custom_data", map[string]interface{}{
	//             "step": "validation",
	//             "input_size": len(someInput),
	//         })
	//
	//         err := complexBusinessLogic()
	//         if err != nil {
	//             hub.CaptureException(err)
	//         }
	//     })
	// }
}

// ExampleErrorHandler shows how to integrate Sentry in your existing error handlers
func ExampleErrorHandler(c *gin.Context) {
	// Simulate an error
	err := fmt.Errorf("example database connection error")

	// Capture the error with all the enriched context
	CaptureError(c, err)

	// Return appropriate response
	c.JSON(500, gin.H{
		"error":   "Internal server error",
		"message": "Something went wrong, please try again later",
	})
}
