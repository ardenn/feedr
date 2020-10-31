package main

import (
	"encoding/xml"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func rssFeed(c echo.Context) error {
	rss := `
   <rss version="2.0">
<channel>
<title>Liftoff News</title>
<link>http://liftoff.msfc.nasa.gov/</link>
<description>Liftoff to Space Exploration.</description>
<language>en-us</language>
<pubDate>Tue, 10 Jun 2003 04:00:00 GMT</pubDate>
<lastBuildDate>Tue, 10 Jun 2003 09:41:01 GMT</lastBuildDate>
<docs>http://blogs.law.harvard.edu/tech/rss</docs>
<generator>Weblog Editor 2.0</generator>
<managingEditor>editor@example.com</managingEditor>
<webMaster>webmaster@example.com</webMaster>
<item>
<title>Star City</title>
<link>http://liftoff.msfc.nasa.gov/news/2003/news-starcity.asp</link>
<description>How do Americans get ready to work with Russians aboard the International Space Station? They take a crash course in culture, language and protocol at Russia's <a href="http://howe.iki.rssi.ru/GCTC/gctc_e.htm">Star City</a>.</description>
<pubDate>Tue, 03 Jun 2003 09:39:21 GMT</pubDate>
<guid>http://liftoff.msfc.nasa.gov/2003/06/03.html#item573</guid>
</item>
<item>
<description>Sky watchers in Europe, Asia, and parts of Alaska and Canada will experience a <a href="http://science.nasa.gov/headlines/y2003/30may_solareclipse.htm">partial eclipse of the Sun</a> on Saturday, May 31st.</description>
<pubDate>Fri, 30 May 2003 11:06:42 GMT</pubDate>
<guid>http://liftoff.msfc.nasa.gov/2003/05/30.html#item572</guid>
</item>
<item>
<title>The Engine That Does More</title>
<link>http://liftoff.msfc.nasa.gov/news/2003/news-VASIMR.asp</link>
<description>Before man travels to Mars, NASA hopes to design new engines that will let us fly through the Solar System more quickly. The proposed VASIMR engine would do that.</description>
<pubDate>Tue, 27 May 2003 08:37:32 GMT</pubDate>
<guid>http://liftoff.msfc.nasa.gov/2003/05/27.html#item571</guid>
</item>
<item>
<title>Astronauts' Dirty Laundry</title>
<link>http://liftoff.msfc.nasa.gov/news/2003/news-laundry.asp</link>
<description>Compared to earlier spacecraft, the International Space Station has many luxuries, but laundry facilities are not one of them. Instead, astronauts have other options.</description>
<pubDate>Tue, 20 May 2003 08:56:02 GMT</pubDate>
<guid>http://liftoff.msfc.nasa.gov/2003/05/20.html#item570</guid>
</item>
</channel>
</rss>
   `
	output := Rss{}
	if err := xml.Unmarshal([]byte(rss), &output); err != nil {
		c.Logger().Error(err)
		return err
	}
	return c.JSON(http.StatusOK, output)
}
func atomFeed(c echo.Context) error {
	atom := `
   <?xml version="1.0" encoding="utf-8"?>
   <feed xmlns="http://www.w3.org/2005/Atom">

     <title>Example Feed</title>
     <link href="http://example.org/"/>
     <updated>2003-12-13T18:30:02Z</updated>
     <author>
       <name>John Doe</name>
     </author>
     <id>urn:uuid:60a76c80-d399-11d9-b93C-0003939e0af6</id>

     <entry>
       <title>Atom-Powered Robots Run Amok</title>
       <link href="http://example.org/2003/12/13/atom03"/>
       <id>urn:uuid:1225c695-cfb8-4ebb-aaaa-80da344efa6a</id>
       <updated>2003-12-13T18:30:02Z</updated>
       <summary>Some text.</summary>
     </entry>

   </feed>
   `
	output := Atom{}
	if err := xml.Unmarshal([]byte(atom), &output); err != nil {
		c.Logger().Error(err)
		return err
	}
	return c.JSON(http.StatusOK, output)
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339_nano} ${level} ${remote_ip} ${method} ${path} ${status} latency: ${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	// Routes
	e.GET("/rss", rssFeed)
	e.GET("/atom", atomFeed)

	// Start server
	e.Logger.Fatal(e.Start(":8000"))
}
