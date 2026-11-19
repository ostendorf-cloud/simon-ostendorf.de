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
          The application serves ready made pages. It writes no access log while doing so. On an
          internal error it writes one line with the request method and a random request identifier,
          without your IP address and without the path you requested.
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
          Umami derives country, region and city from the shortened IP address. This website does
          not store or log your IP address.
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
          This website is operated on my own hardware in Germany, at my home. There is no hosting
          provider and no processor for hosting.
        </p>
        <p>
          The logs of the application and of the ingress in front of it are deleted after 30 days.
          The statistics data in Umami is deleted after 14 months.
        </p>
        <p>
          The legal basis for the logs and for counting page views is Art. 6(1)(f) GDPR. I have a
          legitimate interest in the secure and stable operation of the website and in anonymous
          measurement of its reach.
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
