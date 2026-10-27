const test=require('node:test'),assert=require('node:assert/strict');
const plugin=require('../index.js');
test('Windows names, approvals and native API remain scoped',async()=>{
 process.env.AIDE_COMPUTER_BRIDGE_TOKEN='t'.repeat(32);const tools={};plugin.apply({settings:{allowedApps:['notepad.exe']},tool:t=>tools[t.name]=t});assert.equal(Object.keys(tools).length,6);
 let sent;const original=global.fetch;global.fetch=async(url,options)=>{sent={url,body:JSON.parse(options.body)};return{ok:true,json:async()=>({ok:true})}};
 try{assert.throws(()=>tools.computer_type.handler({text:'x'}),/确认/);await tools.computer_type.handler({text:'中文 Ω',approved:true});assert.equal(sent.body.approved,true);assert.deepEqual(sent.body.allowedApps,['notepad.exe']);assert.equal(sent.body.text,'中文 Ω');assert.equal(sent.url,'http://host.docker.internal:17778/v1/type');assert.throws(()=>plugin._test.bridgeConfig({bridgeUrl:'http://example.com:17778'}));}finally{global.fetch=original;delete process.env.AIDE_COMPUTER_BRIDGE_TOKEN}
});
