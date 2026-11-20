<script setup lang="ts">
import { OPERATOR } from './operator'
</script>

<template>
  <div class="flex flex-col gap-6 text-text">
    <p class="text-text-muted">
      This is a translation. The German version of this page is the legally binding one.
    </p>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">In short</h2>
      <ul class="flex list-disc flex-col gap-2 pl-6">
        <li>This website sets no cookies.</li>
        <li>It loads no scripts, fonts or other content from third party servers.</li>
        <li>There is no form and no login.</li>
        <li>
          Page views are counted on the server, without a script in your browser. How that works is
          described under 2.4.
        </li>
        <li>The website is operated on my own hardware in Germany.</li>
      </ul>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">1. Controller</h2>
      <p>
        {{ OPERATOR.name }}<br />
        {{ OPERATOR.street }}<br />
        {{ OPERATOR.postalCode }} {{ OPERATOR.city }}, {{ OPERATOR.country }}<br />
        Email:
        <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
          OPERATOR.email
        }}</a>
      </p>
    </section>

    <section class="flex flex-col gap-4">
      <h2 class="text-xl font-semibold text-text">
        2. What this website itself stores and processes
      </h2>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.1 Serving the pages</h3>
        <p>
          The application serves ready made pages and writes an access log while doing so, one line
          per request. The line contains:
        </p>
        <ul class="flex list-disc flex-col gap-2 pl-6">
          <li>the time of the request</li>
          <li>the request method</li>
          <li>the page as a route pattern, never the path or the parameters</li>
          <li>the HTTP status code and the amount of data transferred</li>
          <li>the duration of the request</li>
          <li>
            your IP address, logged in full; behind the ingress in front of the application it is
            the address the ingress reports as yours
          </li>
          <li>a random request identifier</li>
        </ul>
        <p>
          The user agent, the referrer and the requested path with its parameters are not logged. On
          an internal error the application additionally writes one line with the request method and
          the request identifier, without your IP address and without the path.
        </p>
        <p>
          Separately, the cluster's ingress in front of the application logs every request in a log
          of its own. It contains your IP address, the date and time, the request line with path and
          parameters, the HTTP status code, the amount of data transferred, the referrer if your
          browser sends it, and the user agent.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.2 Cookies and browser storage</h3>
        <p>
          The website sets no cookies. If you change the appearance with the button in the header
          (light, dark or system setting), your browser keeps that choice locally under the key
          <code>govue-theme</code> in local storage. The entry holds the name of the design and the
          chosen mode. It does not leave your device and is only written when you change the
          appearance. The language follows from the address of the page and is not stored.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.3 Content from other servers</h3>
        <p>
          The website embeds no scripts, fonts, images, maps or videos from third party servers.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.4 Counting page views</h3>
        <p>
          This website counts page views on the server with Umami. No script runs in your browser,
          no cookie is set and nothing is read from or stored on your device. A view is counted
          after an HTML page was served successfully.
        </p>
        <p>The following is sent to Umami:</p>
        <ul class="flex list-disc flex-col gap-2 pl-6">
          <li>
            the page as a route pattern, for example <code>/en/cv/</code>, never the address with
            parameters or a query string
          </li>
          <li>the origin of a referring website (scheme, host, port), never its path</li>
          <li>the first language tag of your browser</li>
          <li>the user agent of your browser</li>
          <li>
            your IP address, shortened before sending to its first three bytes (IPv4, /24) or its
            first six bytes (IPv6, /48)
          </li>
          <li>the host name of this website and the identifier of the website at Umami</li>
        </ul>
        <p>
          Umami derives country, region and city from the shortened IP address. From the shortened
          IP address, the user agent, a secret key and the current month, Umami forms a pseudonymous
          identifier to group together the page views of one visit. The identifier changes every
          month. Umami does not receive the full IP address and does not store the full user agent,
          only the information derived from them (country, region, city, browser, operating system)
          together with the pseudonymous identifier. No merging with other data, and no
          identification of individual visitors, takes place. The application's access log (2.1) and
          the ingress log are separate from this.
        </p>
        <p>
          A visitor who sends the Do Not Track header (<code>DNT: 1</code>) or Global Privacy
          Control (<code>Sec-GPC: 1</code>) is not counted. Also not counted are prefetches, bots,
          requests from private addresses, redirects, error pages, files such as scripts and images,
          <code>robots.txt</code> and <code>sitemap.xml</code>.
        </p>
        <p>
          I operate Umami myself on my own hardware in Germany. The statistics data is deleted after
          14 months.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.5 Contact by email or phone</h3>
        <p>
          If you write to me or call me, I process your details and the content of your message to
          answer your request.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.6 Operation, retention and legal basis</h3>
        <p>
          This website is operated on my own hardware in Germany. There is no hosting provider and
          no processor for hosting.
        </p>
        <p>
          The application's access log and the log of the ingress in front of it are each deleted
          after 30 days. The statistics data in Umami is deleted after 14 months.
        </p>
        <p>
          For both logs the legal basis is Art. 6(1)(f) GDPR. My legitimate interest is the secure
          operation of the website, the analysis of errors and the defence against abuse. I weighed
          this interest against your rights and keep the data minimal.
        </p>
        <p>
          For counting page views the legal basis is Art. 6(1)(f) GDPR. My legitimate interest is to
          understand how the website is used. I weighed this interest against your rights and keep
          the data minimal.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.7 Objection</h3>
        <p>
          You can object to the counting of page views and to the logs at any time (Art. 21 GDPR) by
          writing to me at
          <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
            OPERATOR.email
          }}</a>. You object to the counting automatically when your browser sends Do Not Track or
          Global Privacy Control; the website honours these signals.
        </p>
      </div>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">3. Your rights</h2>
      <p>Under the General Data Protection Regulation you have:</p>
      <ul class="flex list-disc flex-col gap-2 pl-6">
        <li>the right of access to your data (Art. 15 GDPR)</li>
        <li>the right to rectification of inaccurate data (Art. 16 GDPR)</li>
        <li>the right to erasure (Art. 17 GDPR)</li>
        <li>the right to restriction of processing (Art. 18 GDPR)</li>
        <li>the right to data portability (Art. 20 GDPR)</li>
        <li>the right to object to processing (Art. 21 GDPR)</li>
      </ul>
      <p>
        Please contact
        <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
          OPERATOR.email
        }}</a>.
      </p>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">4. Complaint to a supervisory authority</h2>
      <p>
        You have the right to lodge a complaint with a data protection supervisory authority (Art.
        77 GDPR), in particular in the member state of your habitual residence, place of work or
        place of the alleged infringement.
      </p>
    </section>
  </div>
</template>
