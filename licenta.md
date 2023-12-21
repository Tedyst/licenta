Aplicatie pentru verificarea securizarii serverelor de baze de date

- sa aiba suport pentru diverse servere (minim mysql, postgres, mongodb, redis)
- sa testeze parolele default sau inexistente
- eventual sa realizeze un brute force cu niste credentiale cunoscute
- sa poata analiza codebase-urile aplicatiilor/serviciilor sa identifice conexiunile la bd si ip-urile aferente
- sa dea userului posibilitatea sa introduca el ip-urile si porturile
- integrare cu github si alte solutii
- monitorizarea modificarii fisierelor care contin credentialele..asta daca repo e privat..daca e public, de vazut alternative..

My take:

- sa poata verifica daca o baza de date foloseste/nu foloseste SSL. Also, sa poata verifica nivelul de patch/versiunea bd si sa verifice CVE-urile cu acea versiune
- sa poata verifica CVE-urile noi aparute prin API-ul https://nvd.nist.gov/developers/vulnerabilities si sa trimita email-uri la owner / chiar sa opreasca fortat toate build-urile pana se rezolva problema de securtiate la bd (scor de >9.5) pentru a arata cat de grava e problema
- de preferat sa fie cate un modul pentru fiecare baza de date aka sa fie variante diferite de probleme pentru fiecare BD
- sa poti scana configuratia sa nu fie chestii de genul logging de toate query-urile, etc..
- sa poata monitoriza numarul de conexiuni si nivelul de trafic (in caz de infiltrare si export de date masiv)
- secret scanner-ul sa ia repo-urile private la care acces prin integrarile de github/gitlab/etc..
- secret scanning-ul sa ia toate ip-urile/adresele si sa le incerce cu toate parolele (+ parolele de default) pentru brute-forcing. daca exista un match, sa dea alerta ca exista creds in repo
- sa existe un worker (docker container) care se pune in infrastructura clientului si merge ca un fel de "proxy" pentru a testa securitatea bazei de date (aka se presupune ca actorul a compromis deja un service, si vedem daca poate intra asa)
- sa fie folosit hashicorp vault pentru a folosi dynamic credentials pentru DB/chestii critice ale clientilor/a da manage la userii bazei de date
- sa dea alerte daca sunt chestii critice precum permisie sa scrii fisiere prin module (sau daca user-ul de admin are cum sa le dea enable)
- aplicatia se va loga ca user-ul de admin si va testa configuratia/diferite chestii
- aplicatia va vedea user-ul cu care se logheaza aplicatia si va vedea permisiile acestuia
- worker-ul din infra clientului va comunica cu serverul principal prin websockets pentru a primi task-uri
- sa existe un mod prin care sa se vada real-time audit log-ul per baza de date sau pentru toate per-client
- sa fie implementat un syslog server prin care se primesc log-uri de la baze de date (trebuie configurat manual de client), asa am putea primi si audit log-ul
- sa poata fi scanate si Docker Images pentru a vedea daca contin parole hardcodate sau versiuni vulnerabile de baze de date/clienti de baze de date
- sa ia lista de useri de la baza de date si hash-urile parolelor lor si sa vada daca se gasesc in repo/docker(pot fi plaintext sau hash-uri sau base64)
- la docker sa analizeze fiecare layer in parte pentru cazurile in care o cheie este bagata cu ADD . si apoi stearsa cu RUN rm -rf .env (sau ceva de genul)
- sa caute in istoricul din Git pentru parole hardcodate
- sa fie facute atentificare 2fa pentru aplicatie + social login
- sa fie facut totul prin organizatii. Fiecare user o sa aiba o organizatie cu numele lui, dar poate crea si alte organizatii in care baga mai multi useri cu permisii diferite

Postgres:

- `revoke insert, delete, update on all tables in schema public from xxx;` - dont allow user to write to public tables
- `alter default privileges in schema public revoke insert,update,delete on tables from xxx;` - dont allow any default user
- sa nu lasi userii sa escaladeze prin `CREATEROLE`
- sa fie instalat pgAudit

Documentatie:

- https://nvd.nist.gov/developers/vulnerabilities
- https://satoricyber.com/postgres-security/3-pillars-of-postgresql-security/
- https://github.com/wagoodman/dive
- https://devops.stackexchange.com/questions/2731/downloading-docker-images-from-docker-hub-without-using-docker
- https://stackoverflow.com/questions/55386202/how-can-i-use-the-docker-registry-api-to-pull-information-about-a-container-get
- https://gobyexample.com/worker-pools
- https://book.hacktricks.xyz/network-services-pentesting/pentesting-postgresql
- https://docs.particular.net/transports/rabbitmq/delayed-delivery
