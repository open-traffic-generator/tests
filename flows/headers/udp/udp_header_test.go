//go:build all || free || b2b

package udp

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/open-traffic-generator/tests/helpers/otg"
)

func TestUdpHeader(t *testing.T) {
	testConst := map[string]interface{}{
		"pktRate":   int64(50),
		"pktCount":  int32(100),
		"pktSize":   int32(128),
		"txMac":     "00:00:01:01:01:01",
		"rxMac":     "00:00:01:01:01:02",
		"txIp":      "1.1.1.1",
		"rxIp":      "1.1.1.2",
		"txUdpPort": int32(5000),
		"rxUdpPort": int32(6000),
	}

	api := otg.NewOtgApi(t)
	c := udpHeaderConfig(api, testConst)

	api.SetConfig(c)

	api.StartTransmit()

	api.WaitFor(
		func() bool { return udpHeaderMetricsOk(api) },
		&otg.WaitForOpts{FnName: "WaitForFlowMetrics"},
	)
}

func udpHeaderConfig(api *otg.OtgApi, tc map[string]interface{}) gosnappi.Config {
	c := api.Api().NewConfig()
	p1 := c.Ports().Add().SetName("p1").SetLocation(api.TestConfig().OtgPorts[0])
	p2 := c.Ports().Add().SetName("p2").SetLocation(api.TestConfig().OtgPorts[1])

	c.Layer1().Add().
		SetName("ly").
		SetPortNames([]string{p1.Name(), p2.Name()}).
		SetSpeed(gosnappi.Layer1SpeedEnum(api.TestConfig().OtgSpeed))

	f1 := c.Flows().Add().SetName("f1")
	f1.TxRx().Port().
		SetTxName(p1.Name()).
		SetRxName(p2.Name())
	f1.Duration().FixedPackets().SetPackets(tc["pktCount"].(int32))
	f1.Rate().SetPps(tc["pktRate"].(int64))
	f1.Size().SetFixed(tc["pktSize"].(int32))
	f1.Metrics().SetEnable(true)

	eth := f1.Packet().Add().Ethernet()
	eth.Src().SetValue(tc["txMac"].(string))
	eth.Dst().SetValue(tc["rxMac"].(string))

	ip := f1.Packet().Add().Ipv4()
	ip.Src().SetValue(tc["txIp"].(string))
	ip.Dst().SetValue(tc["rxIp"].(string))

	udp := f1.Packet().Add().Udp()
	udp.SrcPort().SetValue(tc["txUdpPort"].(int32))
	udp.DstPort().SetValue(tc["rxUdpPort"].(int32))

	api.Testing().Logf("Config:\n%v\n", c)
	return c
}

func udpHeaderMetricsOk(api *otg.OtgApi) bool {
	m := api.GetFlowMetrics()[0]
	return m.Transmit() == gosnappi.FlowMetricTransmit.STOPPED &&
		m.FramesTx() == 100 &&
		m.FramesRx() == 100
}
