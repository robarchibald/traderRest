package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/robarchibald/configReader"
	"github.com/robarchibald/tpt"
	uuid "github.com/satori/go.uuid"
)

func main() {
	c, err := createClient()
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/api/users/:userID", getUser)
	router.POST("/api/accounts", auth(c, createAccount))
	router.GET("/api/accounts/:accountID", auth(c, getAccount))
	router.PATCH("/api/accounts/:accountID", auth(c, updateAccount))
	router.GET("/api/accounts/:accountID/status", auth(c, getAccountStatus))
	router.GET("/api/accounts", auth(c, getAccounts))
	router.GET("/api/accounts/:accountID/applicants/:applicantID", auth(c, getApplicant))
	router.PATCH("/api/accounts/:accountID/applicants/:applicantID", auth(c, updateApplicant))
	router.GET("/api/accounts/:accountID/beneficiaries/:beneficiaryID", auth(c, getBeneficiary))
	router.PATCH("/api/accounts/:accountID/beneficiaries/:beneficiaryID", auth(c, updateBeneficiary))
	router.POST("/api/files", auth(c, createFile))
	router.GET("/api/files/:fileID", auth(c, getFile))
	router.GET("/api/accounts/:accountID/sources", auth(c, getSources))
	router.POST("/api/accounts/:accountID/sources/:sourceID", auth(c, createSource))
	router.GET("/api/accounts/:accountID/sources/:sourceID", auth(c, getSource))
	router.PATCH("/api/accounts/:accountID/sources/sourceID", auth(c, updateSource))
	router.DELETE("/api/accounts/:accountID/sources/:sourceID", auth(c, removeSource))
	router.POST("/api/accounts/:accountID/sources/:sourceID/verify", auth(c, verifyDeposits))
	router.POST("/api/accounts/:accountID/sources/:sourceID/reverify", auth(c, reverifyDeposits))
	router.GET("/api/accounts/:accountID/transfers", auth(c, getTransfers))
	router.POST("/api/accounts/:accountID/transfers/:transferID", auth(c, initiateTransfer))
	router.GET("/api/accounts/:accountID/transfers/:transferID", auth(c, getTransfer))
	router.DELETE("/api/accounts/:accountID/transfers/:transferID", auth(c, removeTransfer))
	router.GET("/api/accounts/:accountID/portfolio/cash/USD", auth(c, getCashBalance))
	router.GET("/api/accounts/:accountID/portfolio/cash/USD/transactions", auth(c, getTransactionHistory))
	router.GET("/api/accounts/:accountID/portfolio/equities", auth(c, getEquities))
	router.GET("/api/accounts/:accountID/portfolio/equities/:symbol/transactions", auth(c, getSymbolTransactions))
	router.GET("/api/accounts/:accountID/orders", auth(c, getOrders))
	router.POST("/api/accounts/:accountID/orders/:orderID", auth(c, createOrder))
	router.GET("/api/accounts/:accountID/orders/:orderID", auth(c, getOrder))
	router.PATCH("/api/accounts/:accountID/orders/:orderID", auth(c, updateOrder))
	router.DELETE("/api/accounts/:accountID/orders/:orderID", auth(c, removeOrder))
	router.GET("/api/market/hours/:date", auth(c, getMarketHours))
	router.GET("/api/market/symbols/:symbol/quote", auth(c, getSymbolQuote))
	router.GET("/api/market/symbols/:symbol", auth(c, getSymbolQuote2))
	router.GET("/api/market/symbols/:symbol/options", auth(c, getSymbolOptions))
	router.GET("/api/market/symbols/:symbol/timeseries/intraday", auth(c, getIntraday))
	router.GET("/api/market/symbols/:symbol/timeseries/eod", auth(c, getEndOfDayHistory))
	router.GET("/api/market/symbols/:symbol/splits", auth(c, getSplits))
	router.GET("/api/market/symbols/:symbol/dividends", auth(c, getDividends))
	router.GET("/api/market/quote", auth(c, getMultiSymbolQuote))
	router.GET("/api/market/overview", auth(c, getMultiSymbolOverview))
	router.GET("/api/market/symbols/:symbol/company/profile", auth(c, getCompanyProfile))
	router.GET("/api/market/symbols/:symbol/company/financials", auth(c, getFinancials))
	router.GET("/api/market/symbols/:symbol/company/ownership", auth(c, getCompanyOwnership))
	router.GET("/api/market/symbols/:symbol/company/earnings/events", auth(c, getEarningsEvents))
	router.GET("/api/market/symbols/:symbol/company/earnings/surprises", auth(c, getEarningsSurprises))
	router.GET("/api/market/symbols/:symbol/company/ratings", auth(c, getCompanyRatings))
	router.GET("/api/market/symbols/:symbol/company/ratios", auth(c, getRatios))
	router.GET("/api/market/symbols/:symbol/company/news", auth(c, getNews))
	router.GET("/api/accounts/:accountID/documents/confirmations", auth(c, getConfirmations))
	router.GET("/api/accounts/:accountID/documents/statements", auth(c, getStatements))
	router.GET("/api/events", auth(c, getEvents))
	router.POST("/api/events", auth(c, publishEvent))
	router.OPTIONS("/api/*path", options)

	log.Fatal(http.ListenAndServe(":1234", router))
}

func createClient() (*tpt.Client, error) {
	cfg := tpt.Config{}
	err := configReader.ReadFile("trader.conf", &cfg)
	if err != nil {
		return nil, err
	}

	return tpt.NewClient(cfg)
}

func auth(c *tpt.Client, f func(*tpt.Client, httprouter.Params, io.ReadCloser) (string, error)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if c.BearerToken == nil || c.BearerToken.Expiry.Sub(time.Now()) > 0 {
			err := c.OAuth()
			if err != nil {
				outputMessage("", err, w)
				return
			}
		}
		json, err := f(c, p, r.Body)
		outputMessage(json, err, w)
	}
}

func outputMessage(json string, err error, w http.ResponseWriter) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, json)
}

func options(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
}

func getUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	outputMessage(`{"userid": 1234, "name": "Rob Archibald", "picture": "static/images/Rob.jpg"}`, nil, w)
}

func createAccount(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	uuid := uuid.NewV4()
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s", uuid)).Post(b).String()
}

func getAccount(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s", p.ByName("accountID"))).String()
}

func getAccountStatus(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/status", p.ByName("accountID"))).String()
}

func updateAccount(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s", p.ByName("accountID"))).Patch(b).String()
}

func getAccounts(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest("/v1/accounts").String()
}

func getApplicant(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/applicants/%s", p.ByName("accountID"), p.ByName("applicantID"))).String()
}

func updateApplicant(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/applicants/%s", p.ByName("accountID"), p.ByName("applicantID"))).Patch(b).String()
}

func getBeneficiary(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/beneficiaries/%s", p.ByName("accountID"), p.ByName("applicantID"))).String()
}

func updateBeneficiary(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/beneficiaries/%s", p.ByName("accountID"), p.ByName("applicantID"))).Patch(b).String()
}

func createFile(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest("/v1/files").Post(b).String()
}

func getFile(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/files/%s", p.ByName("fileID"))).String()
}

func getSources(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources", p.ByName("accountID"))).String()
}

func createSource(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources/%s", p.ByName("accountID"), p.ByName("sourceID"))).Post(b).String()
}

func getSource(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources/%s", p.ByName("accountID"), p.ByName("sourceID"))).String()
}

func updateSource(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources/sourceID", p.ByName("accountID"))).Patch(b).String()
}

func removeSource(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources/%s", p.ByName("accountID"), p.ByName("sourceID"))).Delete().String()
}

func verifyDeposits(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources/%s/verify", p.ByName("accountID"), p.ByName("sourceID"))).Post(b).String()
}

func reverifyDeposits(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/sources/%s/reverify", p.ByName("accountID"), p.ByName("sourceID"))).Post(b).String()
}

func getTransfers(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/transfers", p.ByName("accountID"))).String()
}

func initiateTransfer(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/transfers/%s", p.ByName("accountID"), p.ByName("transferID"))).Post(b).String()
}

func getTransfer(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/transfers/%s", p.ByName("accountID"), p.ByName("transferID"))).String()
}

func removeTransfer(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/transfers/%s", p.ByName("accountID"), p.ByName("transferID"))).Delete().String()
}

func getCashBalance(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/portfolio/cash/USD", p.ByName("accountID"))).String()
}

func getTransactionHistory(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/portfolio/cash/USD/transactions", p.ByName("accountID"))).String()
}

func getEquities(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/portfolio/equities", p.ByName("accountID"))).String()
}

func getSymbolTransactions(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/portfolio/equities/%s/transactions", p.ByName("accountID"), p.ByName("symbol"))).String()
}

func getOrders(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/orders", p.ByName("accountID"))).String()
}

func createOrder(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/orders/%s", p.ByName("accountID"), p.ByName("orderID"))).Post(b).String()
}

func getOrder(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/orders/%s", p.ByName("accountID"), p.ByName("orderID"))).String()
}

func updateOrder(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/orders/%s", p.ByName("accountID"), p.ByName("orderID"))).Patch(b).String()
}

func removeOrder(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/orders/%s", p.ByName("accountID"), p.ByName("orderID"))).Delete().String()
}

func getMarketHours(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/hours/%s", p.ByName("date"))).String()
}

func getSymbolQuote(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/quote", p.ByName("symbol"))).String()
}

func getSymbolQuote2(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s", p.ByName("symbol"))).String()
}

func getSymbolOptions(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/options", p.ByName("symbol"))).String()
}

func getIntraday(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/timeseries/intraday", p.ByName("symbol"))).String()
}

func getEndOfDayHistory(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/timeseries/eod", p.ByName("symbol"))).String()
}

func getSplits(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/splits", p.ByName("symbol"))).String()
}

func getDividends(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/dividends", p.ByName("symbol"))).String()
}

func getMultiSymbolQuote(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/quote")).String()
}

func getMultiSymbolOverview(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/overview")).String()
}

func getCompanyProfile(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/profile", p.ByName("symbol"))).String()
}

func getFinancials(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/financials", p.ByName("symbol"))).String()
}

func getCompanyOwnership(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/ownership", p.ByName("symbol"))).String()
}

func getEarningsEvents(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/earnings/events", p.ByName("symbol"))).String()
}

func getEarningsSurprises(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/earnings/surprises", p.ByName("symbol"))).String()
}

func getCompanyRatings(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/ratings", p.ByName("symbol"))).String()
}

func getRatios(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/ratios", p.ByName("symbol"))).String()
}

func getNews(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/market/symbols/%s/company/news", p.ByName("symbol"))).String()
}

func getConfirmations(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/documents/confirmations", p.ByName("accountID"))).String()
}

func getStatements(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/accounts/%s/documents/statements", p.ByName("accountID"))).String()
}

func getEvents(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/events")).String()
}

func publishEvent(c *tpt.Client, p httprouter.Params, b io.ReadCloser) (string, error) {
	return c.NewRequest(fmt.Sprintf("/v1/events")).Post(b).String()
}
