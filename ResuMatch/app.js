// document.getElementById("load-vacancies").addEventListener("click", async () => {
//     try {
//         // Отправляем GET-запрос на бэкенд
//         let response = await fetch("http://localhost:8000/vacancies");
//         if (!response.ok) throw new Error(`Ошибка: ${response.status}`);
        
//         let vacancies = await response.json();

//         // Очищаем предыдущие данные
//         let vacancyList = document.getElementById("vacancy-list");
//         vacancyList.innerHTML = "";

//         // Добавляем вакансии в HTML
//         vacancies.forEach(vacancy => {
//             let div = document.createElement("div");
//             div.classList.add("vacancy");
//             div.innerHTML = `
//                 <h2>${vacancy.title}</h2>
//                 <p><strong>Компания:</strong> ${vacancy.company}</p>
//                 <p><strong>Город:</strong> ${vacancy.location}</p>
//                 <p><strong>Зарплата:</strong> ${vacancy.salary}</p>
//                 <p><strong>Описание:</strong> ${vacancy.description}</p>
//                 <p><small>Опубликовано: ${new Date(vacancy.created_at).toLocaleDateString()}</small></p>
//             `;
//             vacancyList.appendChild(div);
//         });

//     } catch (error) {
//         console.error("Ошибка при загрузке вакансий:", error);
//     }
// });


// // Функция для смены страницы без перезагрузки
// function navigate(event, path) {
//     event.preventDefault(); // Останавливаем стандартный переход
//     history.pushState({}, "", path); // Обновляем URL в адресной строке
//     renderPage(); // Отрисовываем нужную страницу
// }

// // Функция загрузки вакансий
// async function loadVacancies() {
//     try {
//         let response = await fetch("http://localhost:8000/vacancies");
//         if (!response.ok) throw new Error(`Ошибка: ${response.status}`);
        
//         let vacancies = await response.json();
//         let content = document.getElementById("content");
//         content.innerHTML = "<h2>Список вакансий</h2>";

//         vacancies.forEach(vacancy => {
//             let div = document.createElement("div");
//             div.classList.add("vacancy");
//             div.innerHTML = `
//                 <h3>${vacancy.title}</h3>
//                 <p><strong>Компания:</strong> ${vacancy.company}</p>
//                 <p><strong>Город:</strong> ${vacancy.location}</p>
//                 <p><strong>Зарплата:</strong> ${vacancy.salary}</p>
//                 <p><strong>Описание:</strong> ${vacancy.description}</p>
//             `;
//             content.appendChild(div);
//         });

//     } catch (error) {
//         console.error("Ошибка загрузки вакансий:", error);
//     }
// }

// // Функция отрисовки контента в зависимости от пути
// function renderPage() {
//     let path = window.location.pathname;
//     let title = document.getElementById("page-title");
//     let content = document.getElementById("content");

//     if (path === "/vacs") {
//         title.textContent = "Вакансии";
//         loadVacancies();
//     } else {
//         title.textContent = "Добро пожаловать";
//         content.innerHTML = "<p>Выберите раздел.</p>";
//     }
// }

// // Обработчик кнопки "Назад" в браузере
// window.onpopstate = renderPage;

// // Загружаем нужную страницу при загрузке
// document.addEventListener("DOMContentLoaded", renderPage);


// Функция смены страницы без перезагрузки
function navigate(event, path) {
    event.preventDefault();
    history.pushState({}, "", path);
    renderPage();
}

// Функция создания карточки вакансии
function createVacancyCard(title, company, location, salary, description) {
    let card = document.createElement("div");
    card.className = "vacancy";
    card.innerHTML = `
        <h3>${title}</h3>
        <p><strong>Компания:</strong> ${company}</p>
        <p><strong>Город:</strong> ${location}</p>
        <p><strong>Зарплата:</strong> ${salary}</p>
        <p><strong>Описание:</strong> ${description}</p>
    `;
    return card;
}

// Функция загрузки вакансий
async function loadVacancies() {
    let vacanciesList = document.getElementById("vacancies-list");
    vacanciesList.innerHTML = ""; // Очищаем список перед загрузкой

    try {
        let response = await fetch("http://localhost:8000/vacancies");
        if (!response.ok) throw new Error(`Ошибка: ${response.status}`);

        let vacancies = await response.json();
        if (vacancies.length === 0) throw new Error("Пустой список вакансий");

        // Создаем карточки для всех вакансий
        vacancies.forEach(vacancy => {
            let card = createVacancyCard(
                vacancy.title, 
                vacancy.company, 
                vacancy.location, 
                vacancy.salary, 
                vacancy.description
            );
            vacanciesList.appendChild(card);
        });

    } catch (error) {
        console.error("Ошибка загрузки вакансий:", error);
        
        // Если бэкенд недоступен или вернул пустой список — создаем три заглушки
        for (let i = 0; i < 3; i++) {
            let card = createVacancyCard("test_error", "test_error", "test_error", "test_error", "test_error");
            vacanciesList.appendChild(card);
        }
    }
}

// Функция отрисовки контента в зависимости от пути
function renderPage() {
    let path = window.location.pathname;
    let title = document.getElementById("page-title");
    let content = document.getElementById("content");
    let vacanciesList = document.getElementById("vacancies-list");

    if (path === "/vacs") {
        title.textContent = "Вакансии";
        content.style.display = "none"; // Скрываем основной контент
        vacanciesList.style.display = "block"; // Показываем список вакансий
        loadVacancies();
    } else {
        title.textContent = "Добро пожаловать";
        content.style.display = "block";
        vacanciesList.style.display = "none"; // Скрываем вакансии
    }
}

// Обработчик кнопки "Назад" в браузере
window.onpopstate = renderPage;

// Загружаем нужную страницу при загрузке
document.addEventListener("DOMContentLoaded", renderPage);
