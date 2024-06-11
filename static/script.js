let currentPage = 1;
let totalPages = 1;
let currentSortColumn = null;
let currentSortOrder = 'asc';

function loadRecords(query) {
    let url = '/api/records?page=' + (currentPage - 1);
    if (query) {
        url += '&search=' + encodeURIComponent(query);
    }
    if (currentSortColumn) {
        url += '&sort=' + encodeURIComponent(currentSortColumn) + '&order=' + encodeURIComponent(currentSortOrder);
    }

    console.log('Fetching records from URL:', url);  // Log URL for debugging
    fetch(url)
        .then(response => response.json())
        .then(data => {
            console.log('Fetched Data:', data);  // Log fetched data

            if (data.hasOwnProperty('records')) {
                const tbody = document.getElementById('dataGrid');
                tbody.innerHTML = '';  // Clear existing rows

                data.records.forEach(record => {
                    const row = document.createElement('tr');
                    document.querySelectorAll('th[data-name]').forEach(th => {
                        const propertyName = th.getAttribute('data-name');
                        const cell = document.createElement('td');
                        cell.textContent = record[propertyName] || '';
                        row.appendChild(cell);
                    });

                    const actionsCell = document.createElement('td');
                    const editButton = document.createElement('button');
                    editButton.textContent = 'Edit';
                    editButton.onclick = () => showEditForm(record['id']);
                    const deleteButton = document.createElement('button');
                    deleteButton.textContent = 'Delete';
                    deleteButton.onclick = () => deleteRecord(record['id']);
                    actionsCell.appendChild(editButton);
                    actionsCell.appendChild(deleteButton);
                    row.appendChild(actionsCell);
                    tbody.appendChild(row);
                });

                currentPage = data.currentPage + 1;
                totalPages = data.totalPages;
                document.getElementById('currentPage').innerText = currentPage;
                document.getElementById('totalPages').innerText = '/ ' + totalPages;
            } else {
                console.error('Data records property is undefined');
            }
        }).catch(error => {
            console.error('Error fetching records:', error);
        });
}

function searchRecords() {
    const query = document.getElementById('searchBox').value;
    currentPage = 1;
    loadRecords(query);
}

// Add event listener to the search box
document.getElementById('searchBox').addEventListener('keyup', function(event) {
    if (event.key === 'Enter') {
        searchRecords();
    }
});

function clearSearchBox() {
    document.getElementById('searchBox').value = '';
    loadRecords();
}

// Add event listener to the search button
document.getElementById('clearSearchInputButton').addEventListener('click', clearSearchBox);

function prevPage() {
    if (currentPage > 1) {
        currentPage--;
        loadRecords();
    }
}

function nextPage() {
    if (currentPage < totalPages) {
        currentPage++;
        loadRecords();
    }
}

function sortRecords(column) {
    if (currentSortColumn === column) {
        currentSortOrder = (currentSortOrder === 'asc') ? 'desc' : 'asc';
    } else {
        currentSortColumn = column;
        currentSortOrder = 'asc';
    }
    currentPage = 1; // Reset to the first page on sort
    loadRecords();
}

function showAddForm() {
    currentId = null;
    document.getElementById('dataForm').reset();
    document.getElementById('modal').style.display = 'block';
}

function showEditForm(id) {
    fetch('/api/records?id=' + id)
        .then(response => response.json())
        .then(data => {
            const record = data.records[0];  // Access the first record
            console.log("Record ==> ", record);
            var modal = document.getElementById('modal');  // Get the modal
            modal.querySelectorAll('input[type="text"]').forEach(input => {      
                const propertyName = input.getAttribute('name'); 
                console.log("propertyName ==> " + propertyName);
                input.value = record[propertyName] || '';
                if (propertyName === "id" ) {
                    console.log("ReadOnly == true");
                    input.setAttribute('readonly', true);  // Make the id field read-only
                    input.setAttribute('disabled', true);  // Make the id field unclickable
                } else {
                    input.removeAttribute('readonly');  // Ensure other fields are not read-only
                    input.removeAttribute('disabled');  // Ensure other fields are not disabled
                }
            }); 
            currentId = id;
            document.getElementById('modal').style.display = 'block';
            
            var inputId = document.getElementById('id');
            console.log(inputId); 


        }).catch(error => {
            console.error('Error fetching record:', error);
        });
}


function hideForm() {
    document.getElementById('modal').style.display = 'none';
}

function submitForm(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const data = {};
    formData.forEach((value, key) => {
        data[key] = value;
    });

    const method = currentId ? 'PUT' : 'POST';
    const url = currentId ? '/api/records?id=' + currentId : '/api/records';

    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(() => {
            loadRecords();
            hideForm();
        })
        .catch(error => {
            console.error('Error submitting form:', error);
        });
}

function deleteRecord(id) {
    if (confirm('Are you sure you want to delete this record?')) {
        fetch('/api/records?id=' + id, { method: 'DELETE' })
            .then(() => loadRecords())
            .catch(error => {
                console.error('Error deleting record:', error);
            });
    }
}

// When the user clicks on <span> (x), close the modal
document.getElementsByClassName('close-button')[0].onclick = function() {
    hideForm();
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
    if (event.target == document.getElementById('modal')) {
        hideForm();
    }
}


document.addEventListener('DOMContentLoaded', () => loadRecords());
