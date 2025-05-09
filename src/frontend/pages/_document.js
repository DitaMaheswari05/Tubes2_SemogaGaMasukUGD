import { Html, Head, Main, NextScript } from 'next/document'

export default function Document() {
  return (
    <Html lang="en">
      <Head>
        <title>Little Alchemy 2 Recipe Finder</title>
        <meta name="description" content="Find recipes for Little Alchemy 2 elements" />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      <body>
        <Main />
        <NextScript />
      </body>
    </Html>
  )
}