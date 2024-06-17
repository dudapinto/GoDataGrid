let currentPage = 1;
let totalPages = 1;
let currentSortColumn = null;
let currentSortOrder = 'asc';

function loadRecords(query) {
    // Remove the existing table to avoid data flickering (old data appearing prior to the new data)
    const tableContainer = document.getElementById('tableContainer');
    tableContainer.innerHTML = '';

    let url = '/api/records?page=' + (currentPage - 1);
    if (query) {
        url += '&search=' + encodeURIComponent(query);
    }
    if (currentSortColumn) {
        url += '&sort=' + encodeURIComponent(currentSortColumn) + '&order=' + encodeURIComponent(currentSortOrder);
    }

    console.log('Fetching records from URL:', url);
    fetch(url)
        .then(response => response.json())
        .then(data => {
            console.log('Fetched Data:', data);

            if (data.columns && data.records) {
                // Create new table
                const table = document.createElement('table');
                table.id = 'dataTable';

                // Create thead and theadRow
                const thead = document.createElement('thead');
                const theadRow = document.createElement('tr');

                // Create header cells for each column
                data.columns.forEach(column => {
                    if (column !== 'id') { // Assuming 'id' is a hidden column
                        const th = document.createElement('th');
                        th.setAttribute('data-name', column);
                        th.setAttribute('data-order', 'asc');
                        th.textContent = capitalizeFirstLetter(column.replace('_', ' ')); // Customize header text as needed
                        th.onclick = () => sortRecords(column);

                        // Create icon element
                        const icon = document.createElement('i');
                        icon.className = 'fas fa-sort'; // Assuming you are using Font Awesome for icons
                        icon.style.marginLeft = '5px'; // Adjust margin as needed
                        th.appendChild(icon);

                        // Update icon based on sorting order
                        if (column === currentSortColumn) {
                            if (currentSortOrder === 'asc') {
                                icon.className = 'fas fa-sort-up'; // Icon for ascending sort
                            } else {
                                icon.className = 'fas fa-sort-down'; // Icon for descending sort
                            }
                        }

                        theadRow.appendChild(th);
                    }
                });

                // Create header cell for Actions
                const thActions = document.createElement('th');
                thActions.textContent = 'Actions';
                theadRow.appendChild(thActions);

                // Append theadRow to thead
                thead.appendChild(theadRow);

                // Append thead to table
                table.appendChild(thead);

                // Create and populate tbody
                const tbody = document.createElement('tbody');
                tbody.id = 'dataGrid';
                data.records.forEach(record => {
                    const row = document.createElement('tr');
                    row.setAttribute('data-id', record['id']); // Store the ID in a data attribute

                    // Create cells for each column
                    data.columns.forEach(column => {
                        if (column !== 'id') {
                            const cell = document.createElement('td');
                            cell.textContent = record[column] || '';
                            row.appendChild(cell);
                        }
                    });

                    // Create actions cell with icons for Edit and Delete
                    const actionsCell = document.createElement('td');

                    const editIcon = document.createElement('i');
                    editIcon.className = 'fas fa-edit'; 
                    editIcon.title = 'Edit'; // Tooltip text for edit action
                    editIcon.onclick = () => showEditForm(record['id']);
                    
                    const deleteIcon = document.createElement('i');
                    deleteIcon.className = 'fas fa-trash-alt'; 
                    deleteIcon.title = 'Delete'; // Tooltip text for delete action
                    deleteIcon.onclick = () => deleteRecordPrompt(record['id']);
                    
                    actionsCell.appendChild(editIcon);
                    actionsCell.appendChild(deleteIcon);
                    row.appendChild(actionsCell);

                    // Append row to tbody
                    tbody.appendChild(row);
                });

                // Replace old table with new table
                table.appendChild(tbody);

                // Append table to the container
                if (tableContainer.firstChild) {
                    tableContainer.replaceChild(table, tableContainer.firstChild);
                } else {
                    tableContainer.appendChild(table);
                }
                table.style.opacity = '1';

                // Update pagination
                currentPage = data.currentPage + 1;
                totalPages = data.totalPages;
                document.getElementById('currentPage').innerText = currentPage;
                document.getElementById('totalPages').innerText = '/ ' + totalPages;

            } else {
                console.error('Data columns or records property is undefined');
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

document.getElementById('searchBox').addEventListener('keyup', function(event) {
    if (event.key === 'Enter') {
        searchRecords();
    }
});

function clearSearchBox() {
    document.getElementById('searchBox').value = '';
    loadRecords();
}

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

    // Update headers to indicate sorting
    const thElements = document.querySelectorAll('#dataTable th');
    thElements.forEach(th => {
        const columnName = th.getAttribute('data-name');
        if (columnName === column) {
            th.setAttribute('data-order', currentSortOrder);
        } else {
            th.setAttribute('data-order', 'asc');
        }
    });

    loadRecords();
}


function showAddForm() {
    currentId = null;
    document.getElementById('dataForm').reset();
    document.getElementById('modal').style.display = 'block';
}

function showEditForm(id) {
    fetch(`/api/records?id=${id}`)
        .then(response => response.json())
        .then(data => {
            const record = data.records[0];
            const modal = document.getElementById('modal');
            const form = document.getElementById('dataForm');

            // Populate form fields
            form.reset();
            for (const key in record) {
                if (record.hasOwnProperty(key)) {
                    const input = form.elements[key];
                    if (input) {
                        input.value = record[key];
                        if (key === 'id') {
                            input.setAttribute('readonly', true);
                            input.setAttribute('disabled', true);
                        } else {
                            input.removeAttribute('readonly');
                            input.removeAttribute('disabled');
                        }
                    }
                }
            }

            currentId = id;
            modal.style.display = 'block';
        })
        .catch(error => {
            console.error('Error fetching record:', error);
        });
}

function hideForm() {
    document.getElementById('modal').style.display = 'none';
}


// Adjust the styling and functionality for the form buttons
document.getElementById('dataForm').addEventListener('submit', submitForm);

function submitForm(event) {
    event.preventDefault();
    const form = event.target;
    const formData = new FormData(form);
    const data = {};
    formData.forEach((value, key) => {
        data[key] = value;
    });

    const method = currentId ? 'PUT' : 'POST';
    const url = currentId ? `/api/records?id=${currentId}` : '/api/records';

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


function showDeleteModal(record) {
    const modal = document.getElementById('deleteModal');
    const deleteDetails = document.getElementById('deleteDetails');
    deleteDetails.innerHTML = '';

    // Display record details
    for (const key in record) {
        if (record.hasOwnProperty(key)) {
            const div = document.createElement('div');
            div.textContent = `${capitalizeFirstLetter(key)}: ${record[key]}`;
            deleteDetails.appendChild(div);
        }
    }

    // Handle delete button click
    document.getElementById('confirmDeleteButton').onclick = function() {
        deleteRecord(record.id);
    };

    modal.style.display = 'block';
}


function hideDeleteModal() {
    document.getElementById('deleteModal').style.display = 'none';
}


function deleteRecord(id) {
    fetch(`/api/records?id=${id}`, { method: 'DELETE' })
        .then(() => {
            loadRecords();
            hideDeleteModal();
        })
        .catch(error => {
            console.error('Error deleting record:', error);
        });
}


function deleteRecordPrompt(id) {
    fetch('/api/records?id=' + id)
        .then(response => response.json())
        .then(data => {
            const record = data.records[0];
            showDeleteModal(record);
        })
        .catch(error => {
            console.error('Error fetching record for deletion:', error);
        });
}

// Ensure cancel button closes the modal
document.getElementById('dataForm').querySelector('button[type="button"]').addEventListener('click', hideForm);


document.getElementsByClassName('close-button')[0].onclick = function() {
    hideForm();
}


window.onclick = function(event) {
    if (event.target == document.getElementById('modal')) {
        hideForm();
    }
}

function capitalizeFirstLetter(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
}

window.addEventListener('pageshow', () => loadRecords());
