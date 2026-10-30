import test from 'node:test'
import assert from 'node:assert/strict'
import {readFile} from 'node:fs/promises'
import ts from 'typescript'

const source=await readFile(new URL('../src/lib/reseller-status.ts',import.meta.url),'utf8')
const output=ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.ES2020}}).outputText
const {peerOperationLabel,peerProvisioningState,resellerAccountLabel}=await import(`data:text/javascript;base64,${Buffer.from(output).toString('base64')}`)
const ready={disabled:false,migration_pending:false,purchase_pending:false,restriction_reasons:[],sync_status:'synced'}

test('account activation is independent of empty prepaid credit',()=>{
 const account={is_active:true,administrative_suspended:false,credit_rial:'0'}
 assert.equal(resellerAccountLabel(account.is_active,account.administrative_suspended),'Active')
 assert.equal(resellerAccountLabel(true,true),'Admin suspended')
 assert.equal(resellerAccountLabel(false,false),'Inactive')
})

for(const [name,patch,label] of [
 ['pending setup',{disabled:true,purchase_pending:true,restriction_reasons:['provisioning'],sync_status:'pending'},'Provisioning'],
 ['failed setup',{disabled:true,purchase_pending:true,provisioning_status:'error',operation_error:'interface: RouterOS returned HTTP 503'},'Provisioning failed'],
 ['paid peer at zero balance',{},'Enabled'],
 ['manual restriction',{disabled:true,manual_disabled:true},'Manually disabled'],
 ['overdue restriction',{disabled:true,restriction_reasons:['billing']},'Restricted'],
 ['admin restriction',{disabled:true,restriction_reasons:['administrative']},'Restricted'],
 ['expired peer',{disabled:true,restriction_reasons:['expired']},'Expired'],
 ['router activation failure',{disabled:true,sync_status:'error'},'Activation failed'],
 ['migration',{migration_pending:true},'Migrating'],
 ['failed free operation',{purchase_pending:true,operation_error:'queue: verification failed'},'Change failed'],
]) test(name,()=>assert.equal(peerOperationLabel({...ready,...patch}),label))

test('older API payload retains pending provisioning state',()=>assert.equal(peerProvisioningState({...ready,restriction_reasons:['provisioning']}),'pending'))
