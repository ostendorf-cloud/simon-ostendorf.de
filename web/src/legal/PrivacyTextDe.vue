<script setup lang="ts">
import { OPERATOR } from './operator'
</script>

<template>
  <div class="flex flex-col gap-6 text-text">
    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">Kurz vorab</h2>
      <ul class="flex list-disc flex-col gap-2 pl-6">
        <li>Diese Website setzt keine Cookies.</li>
        <li>Sie lädt keine Skripte, Schriften oder anderen Inhalte von fremden Servern nach.</li>
        <li>Es gibt kein Formular und keine Anmeldung.</li>
        <li>
          Seitenaufrufe werden auf dem Server gezählt, ohne ein Skript in Ihrem Browser. Wie das
          geht, steht unter 2.4.
        </li>
        <li>Die Website wird auf eigener Hardware in Deutschland betrieben.</li>
      </ul>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">1. Verantwortlicher</h2>
      <p>
        {{ OPERATOR.name }}<br />
        {{ OPERATOR.street }}<br />
        {{ OPERATOR.postalCode }} {{ OPERATOR.city }}<br />
        E-Mail:
        <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
          OPERATOR.email
        }}</a>
      </p>
    </section>

    <section class="flex flex-col gap-4">
      <h2 class="text-xl font-semibold text-text">
        2. Was diese Website selbst speichert und verarbeitet
      </h2>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.1 Auslieferung der Seiten</h3>
        <p>
          Die Anwendung liefert fertige Seiten aus und schreibt dabei ein Zugriffsprotokoll mit
          einer Zeile je Anfrage. Die Zeile enthält:
        </p>
        <ul class="flex list-disc flex-col gap-2 pl-6">
          <li>die Anfragemethode</li>
          <li>die aufgerufene Seite als Routenmuster, nie den Pfad oder die Parameter</li>
          <li>den HTTP-Statuscode und die Menge der übertragenen Daten</li>
          <li>die Dauer der Bearbeitung</li>
          <li>
            Ihre IP-Adresse, die ungekürzt protokolliert wird; hinter dem vorgeschalteten Ingress
            ist es die Adresse, die dieser als Ihre meldet
          </li>
          <li>eine zufällige Anfragekennung</li>
        </ul>
        <p>
          Nicht protokolliert werden der User-Agent, der Referrer und der aufgerufene Pfad mit
          seinen Parametern. Bei einem internen Fehler schreibt die Anwendung zusätzlich eine Zeile
          mit der Anfragemethode und der Anfragekennung, ohne Ihre IP-Adresse und ohne den Pfad.
        </p>
        <p>
          Unabhängig davon protokolliert der Ingress des Clusters, der vor der Anwendung steht,
          jeden Aufruf in einem eigenen Protokoll. Es enthält Ihre IP-Adresse, Datum und Uhrzeit,
          die Anfragezeile mit Pfad und Parametern, den HTTP-Statuscode, die übertragene Datenmenge,
          den Referrer, sofern Ihr Browser ihn übermittelt, und den User-Agent.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.2 Cookies und Speicher im Browser</h3>
        <p>
          Die Website setzt keine Cookies. Wenn Sie über die Schaltfläche in der Kopfzeile die
          Darstellung wechseln (hell, dunkel oder Systemeinstellung), speichert Ihr Browser diese
          Wahl lokal unter dem Schlüssel <code>govue-theme</code> im Local Storage. Der Eintrag
          enthält den Namen des Designs und den gewählten Modus. Er verlässt Ihr Gerät nicht und
          wird nur geschrieben, wenn Sie die Darstellung wechseln. Die Sprache ergibt sich aus der
          Adresse der Seite und wird nicht gespeichert.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.3 Inhalte von anderen Servern</h3>
        <p>
          Die Website bindet keine Skripte, Schriften, Bilder, Karten oder Videos von fremden
          Servern ein.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.4 Zählung der Seitenaufrufe</h3>
        <p>
          Diese Website zählt Seitenaufrufe auf dem Server mit Umami. Dafür läuft kein Skript in
          Ihrem Browser, es wird kein Cookie gesetzt und nichts auf Ihrem Gerät gelesen oder
          gespeichert. Gezählt wird, nachdem eine HTML-Seite erfolgreich ausgeliefert wurde.
        </p>
        <p>An Umami werden folgende Angaben gesendet:</p>
        <ul class="flex list-disc flex-col gap-2 pl-6">
          <li>
            die aufgerufene Seite als Routenmuster, zum Beispiel <code>/de/cv/</code>, nie die
            Adresse mit Parametern oder einem Suchanfrageteil
          </li>
          <li>die Herkunft einer verweisenden Website (Schema, Host, Port), nie ihr Pfad</li>
          <li>die erste Sprachangabe Ihres Browsers</li>
          <li>der User-Agent Ihres Browsers</li>
          <li>
            Ihre IP-Adresse, vor dem Senden auf die ersten drei Bytes (IPv4, /24) oder die ersten
            sechs Bytes (IPv6, /48) gekürzt
          </li>
          <li>der Hostname dieser Website und die Kennung der Website bei Umami</li>
        </ul>
        <p>
          Umami leitet aus der gekürzten IP-Adresse Land, Region und Stadt ab. Aus gekürzter
          IP-Adresse, User-Agent, einem geheimen Schlüssel und dem laufenden Monat bildet Umami eine
          pseudonyme Kennung, um die Seitenaufrufe eines Besuchs zusammenzufassen; die Kennung
          wechselt deshalb jeden Monat. Die vollständige IP-Adresse erhält Umami nicht, und den
          vollständigen User-Agent speichert Umami nicht, sondern nur die daraus abgeleiteten
          Angaben (Land, Region, Stadt, Browser, Betriebssystem) zusammen mit der pseudonymen
          Kennung. Eine Zusammenführung mit anderen Daten oder eine Identifizierung einzelner
          Besucher findet nicht statt. Das Zugriffsprotokoll der Anwendung (2.1) und das Protokoll
          des Ingress sind davon getrennt.
        </p>
        <p>
          Wer den Header Do Not Track (<code>DNT: 1</code>) oder Global Privacy Control (<code
            >Sec-GPC: 1</code
          >) sendet, wird nicht gezählt. Nicht gezählt werden außerdem Vorabrufe, Bots, Anfragen von
          privaten Adressen, Weiterleitungen, Fehlerseiten, Dateien wie Skripte und Bilder sowie
          <code>robots.txt</code> und <code>sitemap.xml</code>.
        </p>
        <p>
          Umami betreibe ich selbst auf eigener Hardware in Deutschland. Die Statistikdaten werden
          nach 14 Monaten gelöscht.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.5 Kontakt per E-Mail oder Telefon</h3>
        <p>
          Wenn Sie mir schreiben oder anrufen, verarbeite ich Ihre Angaben und den Inhalt Ihrer
          Nachricht, um Ihre Anfrage zu beantworten.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">
          2.6 Betrieb, Aufbewahrung und Rechtsgrundlage
        </h3>
        <p>
          Diese Website wird auf eigener Hardware in Deutschland betrieben. Es gibt keinen
          Hosting-Anbieter und keinen Auftragsverarbeiter für das Hosting.
        </p>
        <p>
          Das Zugriffsprotokoll der Anwendung und das Protokoll des vorgeschalteten Ingress werden
          jeweils nach 30 Tagen gelöscht. Die Statistikdaten von Umami werden nach 14 Monaten
          gelöscht.
        </p>
        <p>
          Die Rechtsgrundlage für beide Protokolle und für die Zählung der Seitenaufrufe ist Art. 6
          Abs. 1 lit. f DSGVO.
        </p>
        <p>
          Die Protokolle dienen dem sicheren und stabilen Betrieb der Website, der Analyse von
          Fehlern und der Abwehr von Missbrauch.
        </p>
        <p>
          Die Zählung der Seitenaufrufe dient dazu, zu verstehen, wie die Website genutzt wird, in
          einer pseudonymen Messung ihrer Reichweite.
        </p>
        <p>
          Ich habe dieses Interesse gegen Ihre Rechte und Freiheiten abgewogen und halte die Daten
          knapp.
        </p>
      </div>

      <div class="flex flex-col gap-3">
        <h3 class="text-lg font-semibold text-text">2.7 Widerspruch</h3>
        <p>
          Sie können der Zählung der Seitenaufrufe und den Protokollen jederzeit widersprechen (Art.
          21 DSGVO), indem Sie mir an
          <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
            OPERATOR.email
          }}</a>
          schreiben. Automatisch widersprechen Sie der Zählung, indem Ihr Browser Do Not Track oder
          Global Privacy Control sendet; diese Signale beachtet die Website.
        </p>
      </div>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">3. Ihre Rechte</h2>
      <p>Sie haben nach der Datenschutz-Grundverordnung das Recht auf:</p>
      <ul class="flex list-disc flex-col gap-2 pl-6">
        <li>Auskunft über Ihre verarbeiteten Daten (Art. 15 DSGVO)</li>
        <li>Berichtigung unrichtiger Daten (Art. 16 DSGVO)</li>
        <li>Löschung (Art. 17 DSGVO)</li>
        <li>Einschränkung der Verarbeitung (Art. 18 DSGVO)</li>
        <li>Datenübertragbarkeit (Art. 20 DSGVO)</li>
        <li>Widerspruch gegen die Verarbeitung (Art. 21 DSGVO)</li>
      </ul>
      <p>
        Wenden Sie sich dafür an
        <a class="text-accent-text hover:underline" :href="`mailto:${OPERATOR.email}`">{{
          OPERATOR.email
        }}</a>.
      </p>
    </section>

    <section class="flex flex-col gap-3">
      <h2 class="text-xl font-semibold text-text">4. Beschwerde bei einer Aufsichtsbehörde</h2>
      <p>
        Sie haben das Recht, sich bei einer Datenschutz-Aufsichtsbehörde zu beschweren (Art. 77
        DSGVO), insbesondere in dem Mitgliedstaat Ihres gewöhnlichen Aufenthaltsorts, Ihres
        Arbeitsplatzes oder des Orts des mutmaßlichen Verstoßes.
      </p>
    </section>
  </div>
</template>
