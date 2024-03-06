package accrual

import (
	"encoding/json"
	"fmt"
	"github.com/stsg/gophermart2/internal/config"
	"net/http"
	"net/url"
	"time"

	"github.com/stsg/gophermart2/internal/models"
)

const httpClientTimeout = time.Minute

type Client struct {
	*http.Client
}

func New() Client {
	return Client{
		Client: &http.Client{
			Timeout: httpClientTimeout,
		},
	}
}

func (c *Client) GetOrderInfo(order models.Order) (res models.Order, err error) {
	req, err := c.buildRequest(order.ID)
	if err != nil {
		return
	}

	resp, err := c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	res = order
	err = json.NewDecoder(resp.Body).Decode(&res)
	return
}

func (c *Client) buildRequest(orderID string) (req *http.Request, err error) {
	uri, err := url.Parse(config.Get().AccrualAddress)
	if err != nil {
		return
	}

	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf(uri.String()+"/api/orders/%s", orderID), nil)
	return
}
