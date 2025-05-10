package bot

const (
	decreePattern = `^\d{1,5}/RD/\d{4}$`
	startMessage  = `🌟 *Bun venit la Cetățenie Analyzer\!* 🇷🇴

Cu acest bot poți verifica starea dosarului tău de redobândire a cetățeniei române\. 

_Cum funcționează?_ 🤔
1\. Trimite numărul dosarului în formatul\: *\[număr\]/RD/\[an\]*
   Exemplu\: ` + "`123/RD/2023`" + `
2\. Așteaptă rezultatul
3\. Dacă ai nevoie de ajutor, apasă pe butonul *\"Meniu\"* sau tastează /help

Succes în procesul tău\! 🍀`

	invalidFormat  = "❌ *Format invalid* \n\nTe rog folosește formatul\\: `\\[număr\\]/RD/\\[an\\]`\n\nExemplu\\: `123/RD/2023`"
	searching      = "🔍 _Caut dosarul\\:_ `%s`\n\nTe rog așteaptă puțin\\."
	errorMessage   = "⚠️ *A apărut o eroare\\:* \n\n`%s`\n\nTe rugăm să încerci din nou mai târziu\\."
	unknownState   = "❓ *Stare necunoscută*\n\nTe rugăm să încerci mai târziu sau să contactezi administratorul\\."
	successMessage = "🎉 *Felicitări\\!* \n\nDosarul `%s` a fost *găsit și rezolvat*\\.\n\n" +
		"Timp preluare date\\: %s\n" +
		"Timp analiză document\\: %s\n\n" +
		"Poți continua cu procedurile ulterioare pentru redobândirea cetățeniei române\\."

	inProgressMsg = "⏳ *Dosar în procesare* \n\nDosarul `%s` a fost *găsit dar nu este rezolvat încă*\\.\n\n" +
		"Timp preluare date\\: %s\n" +
		"Timp analiză document\\: %s\n\n" +
		"Va trebui să mai aștepți până când va fi finalizat\\."

	notFoundMsg = "🔎 *Rezultat negativ* \n\nDosarul `%s` *nu a fost găsit*\\.\n\n" +
		"Timp preluare date\\: %s\n" +
		"Timp analiză document\\: %s\n\n" +
		"Te rugăm să verifici numărul și anul\\, sau să contactezi autoritățile competente\\."

	helpMessage = `ℹ️ *Ajutor și instrucțiuni*

📌 _Cum verific dosarul?_
Trimite numărul dosarului în formatul\: *\[număr\]/RD/\[an\]*
Exemplu\: ` + "`123/RD/2023`" + `

📌 _Ce înseamnă rezultatele?_
✅ *Găsit și rezolvat* \- Dosar finalizat, poți continua procedurile
🔄 *Găsit dar nerezolvat* \- Dosar în procesare, mai așteaptă
❌ *Negăsit* \- Verifică numărul sau contactează autoritățile

📌 _Comenzi disponibile\:_
/start \- Mesaj de bun venit
/help \- Acest mesaj de ajutor`
)
