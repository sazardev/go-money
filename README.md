# GO Money - A very simple CLI money management application extracting all your emails from Gmail and summarizing your expenses.

## Overview

GO Money is a command-line interface (CLI) application written in Go that helps you manage your finances by extracting transaction data from your Gmail account. It scans your emails for purchase receipts and summarizes your expenses, providing insights into your spending habits.

Also, make a graph using Go Echarts to visualize your expenses over time, making it easier to understand your financial patterns.

Go Money can make your expenses from:

- Transportation Services as: Uber, Lyft, Didi, Grab, InDriver, etc.
- Food Delivery Services as: Uber Eats, DoorDash, Grubhub, Postmates, Deliveroo, Rappi, DiDi Food, etc.
- E-commerce Platforms as: Amazon, eBay, AliExpress, Etsy, Walmart, Sams Club, Google Shopping, Play Store, Apple Store, etc.
- Subscription Services as: Netflix, Spotify, Disney+, Hulu, Apple Music, HBO Max, Amazon Prime, Youtube Premium, Youtube Music, etc.
- Travel and Accommodation Services as: Airbnb, Booking.com, Expedia, etc.
- Financial Services as: PayPal, Venmo, Cash App, Bank transfer receipts, Banorte, Banamex, BBVA, HSBC, Mercado Pago, Nubank, etc.
- Cinema Services as: Fandango, AMC Theatres, Cinemex, Cinepolis, etc.
- Video Game Platforms as: Steam, Epic Games Store, Xbox, PlayStation Store, etc.
- Clothing and Retail Services as: Zara, H&M, Nike, Adidas, etc.
- And many more!

## Features

- Extracts transaction data from Gmail purchase receipts.
- Summarizes expenses by category and time period.
- Generates graphical representations of expenses using Go Echarts.
- Easy-to-use CLI interface.
- Secure OAuth2 authentication with Google.
- Lightweight and fast performance.
- Open-source and customizable.
- Cross-platform compatibility (Windows, macOS, Linux).
- Regular updates and improvements.
- Concurrent processing for faster email scanning.
- Detailed logging and error handling.
- Configurable settings for personalized expense tracking.
- High accuracy in data extraction using advanced parsing techniques.
- CSV export of expense summaries for further analysis.
- Modular architecture for easy integration with other tools.
- Comprehensive documentation and user guides.
- Active community support and contributions.

## Usage

Go money needs to generate OAuth2 credentials to access your Gmail account. You just need to use:

```bash
gm auth login
```

This command will open a browser window where you can log in to your Google account and authorize the application to access your Gmail data.

After successfully logging in, you can run the following command to extract and summarize your expenses:

```bash
gm calculate
```

This command will scan your Gmail account for purchase receipts, extract the relevant transaction data, and provide a summary of your expenses.

You can also generate a graphical representation of your expenses using:

```bash
gm graph
```

This command will create a graph visualizing your expenses over time using Go Echarts.

# Commands

- `gm auth login`: Authenticate with your Google account using OAuth2.
- `gm calculate`: Extract and summarize your expenses from Gmail purchase receipts.
- `gm graph`: Generate a graphical representation of your expenses using Go Echarts.
- `gm help`: Display help information about the available commands.
- `gm version`: Show the current version of the GO Money application.