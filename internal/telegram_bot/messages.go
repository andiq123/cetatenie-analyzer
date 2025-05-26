package telegram_bot

const (
	decreePattern = `^\d{1,5}/RD/\d{4}$`
	startMessage  = `🌟 <b>Bun venit la Cetățenie Analyzer!</b> 🇷🇴

Cu acest bot poți verifica starea dosarului tău de redobândire a cetățeniei române și să primești notificări când se schimbă starea. 

<i>Cum funcționează?</i> 🤔
1️⃣ Trimite numărul dosarului în formatul: <b>[număr]/RD/[an]</b>
   Exemplu: <code>123/RD/2023</code>
2️⃣ Așteaptă rezultatul
3️⃣ Poți adăuga dosarul la notificări pentru a fi anunțat când se schimbă starea
4️⃣ Pentru ajutor, tastează /ajutor

Succes în procesul tău! 🍀`

	invalidFormat  = "❌ <b>Format invalid</b>\n\nTe rog folosește formatul: <b>[număr]/RD/[an]</b>\nExemplu: <code>123/RD/2023</code>"
	searching      = "🔍 <b>Căutare în curs...</b>\n\nDosar: <code>%s</code>\n\nTe rog așteaptă puțin."
	errorMessage   = "⚠️ <b>A apărut o eroare</b>\n\n<code>%s</code>\n\nTe rugăm să încerci din nou mai târziu."
	unknownState   = "❓ <b>Stare necunoscută</b>\n\nTe rugăm să încerci mai târziu sau să contactezi administratorul."
	successMessage = "🎉 <b>Felicitări!</b>\n\nDosarul <code>%s</code> a fost <b>găsit și rezolvat</b>.\n\n" +
		"⏱️ Timp preluare date: %s\n" +
		"⏱️ Timp analiză document: %s\n\n" +
		"Poți continua cu procedurile ulterioare pentru redobândirea cetățeniei române."

	inProgressMsg = "⏳ <b>Dosar în procesare</b>\n\nDosarul <code>%s</code> a fost <b>găsit dar nu este rezolvat încă</b>.\n\n" +
		"⏱️ Timp preluare date: %s\n" +
		"⏱️ Timp analiză document: %s\n\n" +
		"Va trebui să mai aștepți până când va fi finalizat."

	notFoundMsg = "🔎 <b>Rezultat negativ</b>\n\nDosarul <code>%s</code> <b>nu a fost găsit</b>.\n\n" +
		"⏱️ Timp preluare date: %s\n" +
		"⏱️ Timp analiză document: %s\n\n" +
		"Te rugăm să verifici numărul și anul, sau să contactezi autoritățile competente."

	helpMessage = "ℹ️ <b>Ajutor și instrucțiuni</b>\n\n" +
		"📌 <b>Cum verific dosarul?</b>\n" +
		"Trimite numărul dosarului în formatul: <b>[număr]/RD/[an]</b>\n" +
		"Exemplu: <code>123/RD/2023</code>\n\n" +
		"📌 <b>Ce înseamnă rezultatele?</b>\n" +
		"✅ <b>Găsit și rezolvat</b> - Dosar finalizat, poți continua procedurile\n" +
		"🔄 <b>Găsit dar nerezolvat</b> - Dosar în procesare, mai așteaptă\n" +
		"❌ <b>Negăsit</b> - Verifică numărul sau contactează autoritățile\n\n" +
		"📌 <b>Comenzi disponibile:</b>\n" +
		"• /start - Pornire bot și mesaj de bun venit\n" +
		"• /ajutor - Ajutor și informații despre comenzi\n" +
		"• /abonamente - Listează toate abonamentele tale\n" +
		"• /adauga [număr]/RD/[an] - Adaugă un abonament la un dosar\n" +
		"   Exemplu: <code>/adauga 123/RD/2023</code>\n" +
		"• /sterge [număr]/RD/[an] - Șterge un abonament la un dosar\n" +
		"   Exemplu: <code>/sterge 123/RD/2023</code>\n" +
		"• /sterge_toate - Șterge toate abonamentele\n\n" +
		"📌 <b>Despre notificări</b>\n" +
		"• Vei primi notificări când starea dosarului se schimbă\n" +
		"• Poți avea mai multe dosare în abonamente\n" +
		"• Notificările sunt trimise automat când se detectează schimbări"
)
