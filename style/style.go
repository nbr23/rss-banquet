package style

import "strings"

func InjectRssStyle(x string) string {
	return strings.Replace(x, `<?xml version="1.0" encoding="UTF-8"?>`, "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<?xml-stylesheet type=\"text/xsl\" href=\"/rss-style.xsl\"?>\n", 1)
}

func InjectAtomStyle(x string) string {
	return strings.Replace(x, `<?xml version="1.0" encoding="UTF-8"?>`, "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<?xml-stylesheet type=\"text/xsl\" href=\"/atom-style.xsl\"?>\n", 1)
}

var RssStyle = `<xsl:stylesheet
                xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:fo="http://www.w3.org/1999/XSL/Format"
                version="1.0">
  <xsl:output method="html"/>
  <xsl:template match="/">
    <html xmlns="http://www.w3.org/1999/xhtml" lang="en">
      <head>
        <title>
          <xsl:value-of select="/rss/channel/title"/> | RSS-banquet
        </title>
        <style>
body {
    font-family: Arial, sans-serif;
    background-color: #f4f4f4;
    margin: 0;
}

.feed-banner {
  background-color: #ffc107;
  color: #000;
  text-align: center;
  padding: 5px;
  margin: 0;
  border-bottom: 2px solid #e0a800;
  border-radius: 0 0 10px 10px;
  font-style: italic;
  font-family: monospace, monospace;
  font-weight: bold;
  font-size: 0.8em;
}

.feed-content {
  margin: 0;
  padding: 20px;
}

.feed-header {
    display: flex;
    align-items: center;
    margin-bottom: 20px;
}

.feed-title {
    font-size: 2em;
    margin: 0;
    color: #333;

    a {
        font-size: 0.5em;
        margin-left: 10px;
        text-decoration: none;
    }
}

.item-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.item {
    display: flex;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    transition: transform 0.2s;
}

.item:hover {
    transform: scale(1.02);
}

.details {
    padding: 20px;
    flex: 1;
    text-decoration: none;
    color: inherit;
}

.title {
    font-size: 1.7em;
    margin: 0 0 10px;
    color: #333;
    font-weight: bold;
}

.content {
    font-size: 1em;
    margin: 0;
    color: #666;
}

.thumbnail {
    width: 200px;
    height: 200px;
    object-fit: scale-down;
}
        </style>
      </head>
      <body>
        <div class="feed-banner">
            This page is an RSS Feed, add it to your feed reader!
        </div>
        <div class="feed-content">
          <div class="feed-header">
            <h1 class="feed-title">
              <xsl:value-of select="/rss/channel/title"/>
              <a target="_blank" rel="noopener noreferrer">
              <xsl:attribute name="href">
                <xsl:value-of select="/rss/channel/link"/>
              </xsl:attribute>
              🔗
              </a>
            </h1>
          </div>
          <div class="item-list">
            <xsl:for-each select="/rss/channel/item">
            <div class="item">
            <xsl:if test="starts-with(enclosure/@type, 'image') and enclosure/@url != ''">
                <img class="thumbnail">
                  <xsl:attribute name="src">
                    <xsl:value-of select="enclosure/@url"/>
                  </xsl:attribute>
                </img>
              </xsl:if>
              <a class="details" target="_blank" rel="noopener noreferrer">
                <xsl:attribute name="href">
                  <xsl:value-of select="link"/>
                </xsl:attribute>
                <h3 class="title"><xsl:value-of select="title"/></h3>
                <p class="content">
                  <xsl:value-of select="description"/>
                </p>
              </a>
            </div>
            </xsl:for-each>
          </div>
        </div>
      </body>
    </html>
  </xsl:template>
</xsl:stylesheet>
`

var AtomStyle = `<xsl:stylesheet
								xmlns:atom="http://www.w3.org/2005/Atom"
                xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:fo="http://www.w3.org/1999/XSL/Format"
                version="1.0">
  <xsl:output method="html"/>
  <xsl:template match="/atom:feed">
    <html xmlns="http://www.w3.org/1999/xhtml" lang="en">
      <head>
        <title>
          <xsl:value-of select="atom:title"/> | RSS-banquet
        </title>
        <style>
body {
    font-family: Arial, sans-serif;
    background-color: #f4f4f4;
    margin: 0;
}

.feed-banner {
  background-color: #ffc107;
  color: #000;
  text-align: center;
  padding: 5px;
  margin: 0;
  border-bottom: 2px solid #e0a800;
  border-radius: 0 0 10px 10px;
  font-style: italic;
  font-family: monospace, monospace;
  font-weight: bold;
  font-size: 0.8em;
}

.feed-content {
  margin: 0;
  padding: 20px;
}

.feed-header {
    display: flex;
    align-items: center;
    margin-bottom: 20px;
}

.feed-title {
    font-size: 2em;
    margin: 0;
    color: #333;

    a {
        font-size: 0.5em;
        margin-left: 10px;
        text-decoration: none;
    }
}

.item-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.item {
    display: flex;
    background-color: #fff;
    border-radius: 8px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    transition: transform 0.2s;
}

.item:hover {
    transform: scale(1.02);
}

.details {
    padding: 20px;
    flex: 1;
    text-decoration: none;
    color: inherit;
}

.title {
    font-size: 1.7em;
    margin: 0 0 10px;
    color: #333;
    font-weight: bold;
}

.content {
    font-size: 1em;
    margin: 0;
    color: #666;
}

.thumbnail {
    width: 200px;
    height: 200px;
    object-fit: scale-down;
}
        </style>
      </head>
      <body>
        <div class="feed-banner">
          This page is an Atom Feed, add it to your feed reader!
        </div>
        <div class="feed-content">
          <div class="feed-header">
            <h1 class="feed-title">
              <xsl:value-of select="atom:title"/>
              <a target="_blank" rel="noopener noreferrer">
              <xsl:attribute name="href">
                <xsl:value-of select="atom:link/@href"/>
              </xsl:attribute>
              🔗
              </a>
            </h1>
          </div>
          <div class="item-list">
            <xsl:for-each select="atom:entry">
            <div class="item">
            <xsl:if test="starts-with(atom:link[@rel='enclosure']/@type, 'image') and atom:link[@rel='enclosure']/@href != ''">
                <img class="thumbnail">
                  <xsl:attribute name="src">
                    <xsl:value-of select="atom:link[@rel='enclosure']/@href"/>
                  </xsl:attribute>
                </img>
              </xsl:if>
              <a class="details" target="_blank" rel="noopener noreferrer">
                <xsl:attribute name="href">
                  <xsl:value-of select="atom:link/@href"/>
                </xsl:attribute>
                <h3 class="title"><xsl:value-of select="atom:title"/></h3>
                <p class="content">
                  <xsl:value-of select="atom:summary"/>
                </p>
              </a>
            </div>
            </xsl:for-each>
          </div>
        </div>
      </body>
    </html>
  </xsl:template>
</xsl:stylesheet>
`
