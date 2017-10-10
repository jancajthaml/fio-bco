let axios = require("axios");
let sync = require("./sync.js");
let options = require("config").get("core");

function Tenant (tenantName) {
  if (!tenantName) {
    throw Error("When creating Tenant you have to provide his name");
  }
  this._tenantName = tenantName;
}

Tenant.prototype._runInParallel = async function (items, parallelismSize, processItem, afterBatch) {
  for (let i = 0; i < items.length; i += parallelismSize) {
    let bulk = items.slice(i, Math.min(i + parallelismSize, items.length));
    let batchResult = await Promise.all(bulk.map((item, index) => processItem(item, index)));
    if (afterBatch) await afterBatch(batchResult);
    console.log("Finished " + (Math.floor(i/parallelismSize)+1) + ". bulk");
  }
};

Tenant.prototype._getApiUrl = function () {
  return options.url + "/v1/" + this._tenantName + "/core";
};

Tenant.prototype.createMissingAccounts = async function (accounts) {
  await this._runInParallel(accounts, options.accountsParallelismSize,
    async account => {
      try {
        await axios.get(this._getApiUrl() + "/account/" + account.accountNumber);
        console.log("Account " + account.accountNumber + " already exists");
      } catch (error) {
        if (error.response && error.response.status === 404) {
          await axios.put(this._getApiUrl() + "/account/", account);
          console.log("Created account " + account.accountNumber);
        } else {
          throw error;
        }
      }
    }
  );
};

Tenant.prototype.createTransactions = async function (transactions, accountNumber) {
  await this._runInParallel(transactions, options.transactionsParallelismSize,
    async (transaction, index) => {
      await axios.put(this._getApiUrl() + "/transaction", transaction);
      // ToDo - add real transaction id
      let transactionId = index;
      console.log("Created transaction ID " + transactionId);
      return transactionId;
    },
    async (transactionIds) => {
      let max = transactionIds.reduce((max, transactionId) => {
        return Math.max(max, transactionId);
      });
      await sync.setTransactionCheckpoint(options.db, accountNumber, max);
      console.log("Max ID " + max);
    }
  );
};

Tenant.prototype.getTransactionCheckpoint = async function (accountNumber) {
  return await sync.getTransactionCheckpoint(options.db, accountNumber);
};

module.exports = {
  Tenant
};