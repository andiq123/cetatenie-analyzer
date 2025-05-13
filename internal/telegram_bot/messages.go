package telegram_bot

const (
	decreePattern = `^\d{1,5}/RD/\d{4}$`
	startMessage  = `🌟 <b>Bun venit la Cetățenie Analyzer!</b> 🇷🇴

Cu acest bot poți verifica starea dosarului tău de redobândire a cetățeniei române și să primești notificări când se schimbă starea. 

<i>Cum funcționează?</i> 🤔
1. Trimite numărul dosarului în formatul: <b>[număr]/RD/[an]</b>
   Exemplu: <code>123/RD/2023</code>
2. Așteaptă rezultatul
3. Poți adăuga dosarul la notificări pentru a fi anunțat când se schimbă starea
4. Pentru ajutor, tastează /ajutor

Succes în procesul tău! 🍀`

	invalidFormat  = "❌ <b>Format invalid</b> \n\nTe rog folosește formatul: <code>[număr]/RD/[an]</code>\n\nExemplu: <code>123/RD/2023</code>"
	searching      = "🔍 <i>Caut dosarul:</i> <code>%s</code>\n\nTe rog așteaptă puțin."
	errorMessage   = "⚠️ <b>A apărut o eroare:</b> \n\n<code>%s</code>\n\nTe rugăm să încerci din nou mai târziu."
	unknownState   = "❓ <b>Stare necunoscută</b>\n\nTe rugăm să încerci mai târziu sau să contactezi administratorul."
	successMessage = "🎉 <b>Felicitări!</b> \n\nDosarul <code>%s</code> a fost <b>găsit și rezolvat</b>.\n\n" +
		"Timp preluare date: %s\n" +
		"Timp analiză document: %s\n\n" +
		"Poți continua cu procedurile ulterioare pentru redobândirea cetățeniei române."

	inProgressMsg = "⏳ <b>Dosar în procesare</b> \n\nDosarul <code>%s</code> a fost <b>găsit dar nu este rezolvat încă</b>.\n\n" +
		"Timp preluare date: %s\n" +
		"Timp analiză document: %s\n\n" +
		"Va trebui să mai aștepți până când va fi finalizat."

	notFoundMsg = "🔎 <b>Rezultat negativ</b> \n\nDosarul <code>%s</code> <b>nu a fost găsit</b>.\n\n" +
		"Timp preluare date: %s\n" +
		"Timp analiză document: %s\n\n" +
		"Te rugăm să verifici numărul și anul, sau să contactezi autoritățile competente."

	helpMessage = "ℹ️ <b>Ajutor și instrucțiuni</b>\n\n" +
		"📌 <i>Cum verific dosarul?</i>\n" +
		"Trimite numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\n" +
		"Exemplu: <code>123/RD/2023</code>\n\n" +
		"📌 <i>Ce înseamnă rezultatele?</i>\n" +
		"✅ <b>Găsit și rezolvat</b> - Dosar finalizat, poți continua procedurile\n" +
		"🔄 <b>Găsit dar nerezolvat</b> - Dosar în procesare, mai așteaptă\n" +
		"❌ <b>Negăsit</b> - Verifică numărul sau contactează autoritățile\n\n" +
		"📌 <i>Comenzi disponibile:</i>\n" +
		"/start - Mesaj de bun venit\n" +
		"/ajutor - Acest mesaj de ajutor\n" +
		"/abonamente - Vezi toate abonamentele tale\n" +
		"/adaugaAbonament [număr] - Adaugă un dosar la notificări\n" +
		"/stergeAbonament [număr] - Șterge un abonament\n" +
		"/stergeToateAbonamentele - Șterge toate abonamentele\n\n" +
		"📌 <i>Despre notificări</i>\n" +
		"• Vei primi notificări când starea dosarului se schimbă\n" +
		"• Poți avea mai multe dosare în abonamente\n" +
		"• Notificările sunt trimise automat când se detectează schimbări"
)
